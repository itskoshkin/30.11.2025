package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"

	apiModels "link-availability-checker/internal/api/models"
	"link-availability-checker/internal/config"
	"link-availability-checker/internal/models"
	"link-availability-checker/internal/storage"
	"link-availability-checker/internal/utils/files"
	"link-availability-checker/pkg/pdf"
)

type LinkService interface {
	CheckLinkSet(links *apiModels.CheckLinkSetRequest) (*models.Set, error)
	GetLinkSetAsPDF(ctx context.Context, set []int) (string, error)
	Shutdown(ctx context.Context) error
}

type linkTask struct {
	set        *models.Set
	resultChan chan *models.Set
}

type fileTask struct {
	Set *models.Set `json:"set"`
}

var ErrServiceStopping = errors.New("service is shutting down, task queued for restart")

type LinkServiceImpl struct {
	ls storage.LinkStorage
	as AvailabilityService

	wg          sync.WaitGroup
	queue       chan *linkTask
	queueMutex  sync.Mutex
	queueClosed bool

	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

func NewLinkService(ls storage.LinkStorage, as AvailabilityService) LinkService {
	ctx, cancel := context.WithCancel(context.Background())

	svc := &LinkServiceImpl{
		ls:             ls,
		as:             as,
		queue:          make(chan *linkTask, 1000),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}

	if err := svc.LoadQueueFromFile(); err != nil {
		log.Printf("Warning: failed to load queue from file: %v", err)
	}

	for i := 0; i < viper.GetInt(config.QueueWorkers); i++ {
		svc.wg.Add(1)
		go svc.worker()
	}

	return svc
}

func (svc *LinkServiceImpl) CheckLinkSet(links *apiModels.CheckLinkSetRequest) (*models.Set, error) {
	set := models.Set{Links: links.ConvertLinksToModel()}

	resultChan := make(chan *models.Set)
	task := &linkTask{set: &set, resultChan: resultChan}

	svc.queueMutex.Lock()
	if svc.queueClosed {
		defer svc.queueMutex.Unlock()
		err := svc.appendTaskToFile(task)
		if err != nil {
			return nil, fmt.Errorf("service is shutting down and failed to save task: %w", err)
		}
		return nil, ErrServiceStopping
	}
	svc.queue <- task
	svc.queueMutex.Unlock()

	return <-resultChan, nil
}

func (svc *LinkServiceImpl) GetLinkSetAsPDF(ctx context.Context, nums []int) (string, error) {
	sets := make([]models.Set, 0, len(nums))

	for _, num := range nums {
		set, err := svc.ls.GetLinkSet(num)
		if err != nil {
			return "", fmt.Errorf("failed to get link set %d: %w", num, err) //TODO: Check custom error not lost here
		}

		if viper.GetBool(config.RecheckStatusesWhenPrinting) {
			domains := make([]string, len(set.Links))
			for i, link := range set.Links {
				domains[i] = link.Domain
			}

			var statuses []bool
			statuses, err = svc.checkDomainsAvailability(ctx, domains)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return "", err
				}
				return "", fmt.Errorf("failed to check domains for set %d: %w", num, err)
			}

			for i, status := range statuses {
				set.Links[i].Status = status
			}
		}

		sets = append(sets, *set)
	}

	text := make([][]string, len(sets))
	for i, set := range sets {
		for j, link := range set.Links {
			statusStr := models.ConvertStatusToString(link.Status)
			text[i] = append(text[i], fmt.Sprintf("%d. %-42s - %s\n", j+1, link.Domain, statusStr))
		}
	}

	filePath, err := pdf.GeneratePDF(nums, text)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return filePath, nil
}

func (svc *LinkServiceImpl) worker() {
	defer svc.wg.Done()

	for task := range svc.queue {
		domains := make([]string, len(task.set.Links))
		for i, link := range task.set.Links {
			domains[i] = link.Domain
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		statuses, err := svc.checkDomainsAvailability(ctx, domains)
		cancel()
		if err != nil && errors.Is(err, context.Canceled) {
			log.Println("Worker: Task processing canceled due to shutdown.")
			continue
		}

		for i, status := range statuses {
			task.set.Links[i].Status = status
		}

		num, _ := svc.ls.SaveLinkSet(task.set)
		task.set.Number = num

		if task.resultChan != nil {
			select {
			case task.resultChan <- task.set:
				// Sent
			default:
				// Skip
				//log.Println("Worker: Result channel send skipped (client stopped listening).")
			}
		}
	}
}

func (svc *LinkServiceImpl) checkDomainsAvailability(ctx context.Context, domains []string) ([]bool, error) {
	type result struct {
		index  int
		status bool
		err    error
	}

	jobs := make(chan int, len(domains))
	results := make(chan result, len(domains))

	numWorkers := len(domains) / viper.GetInt(config.WorkersRatio) // Config validation enforces WorkersRatio > 0
	if numWorkers > viper.GetInt(config.MaxWorkers) {
		numWorkers = viper.GetInt(config.MaxWorkers) // Hard limit
	}

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				status, err := svc.as.CheckDomainAvailability(ctx, domains[i])
				results <- result{index: i, status: status, err: err}
			}
		}()
	}

	for i := range domains {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	close(results)

	statuses := make([]bool, len(domains))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		statuses[res.index] = res.status
	}

	return statuses, nil
}

func (svc *LinkServiceImpl) LoadQueueFromFile() error {
	svc.queueMutex.Lock() //MARK: Needed?
	defer svc.queueMutex.Unlock()

	filePath := viper.GetString(config.QueueFilePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Assume no file means no queued tasks
		}
		return err
	}

	var fileTasks []fileTask
	if err = json.Unmarshal(data, &fileTasks); err != nil {
		return err
	}

	log.Printf("[SERVICE] Loading remaining %d tasks from queue...", len(fileTasks))

	for _, ft := range fileTasks {
		task := &linkTask{
			set:        ft.Set,
			resultChan: nil, // No need to return results for loaded tasks after restart - no clients waiting
		}
		select {
		case svc.queue <- task:
		default:
			log.Println("Queue is full, skipping restored task")
		}
	}

	files.Delete(filePath)
	return nil
}

func (svc *LinkServiceImpl) drainQueue() []*linkTask {
	tasks := make([]*linkTask, 0, len(svc.queue))
	for task := range svc.queue {
		tasks = append(tasks, task)
	}
	return tasks
}

func (svc *LinkServiceImpl) SaveQueueToFile(tasks []*linkTask) error {
	fileTasks := make([]fileTask, len(tasks))
	for i, t := range tasks {
		fileTasks[i] = fileTask{Set: t.set}
	}

	data, err := json.MarshalIndent(fileTasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(viper.GetString(config.QueueFilePath), data, 0644)
}

func (svc *LinkServiceImpl) appendTaskToFile(task *linkTask) error {
	var fileTasks []fileTask

	data, err := os.ReadFile(viper.GetString(config.QueueFilePath))
	if err == nil {
		_ = json.Unmarshal(data, &fileTasks)
	}

	fileTasks = append(fileTasks, fileTask{Set: task.set})

	q, err := json.MarshalIndent(fileTasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(viper.GetString(config.QueueFilePath), q, 0644)
}

func (svc *LinkServiceImpl) Shutdown(ctx context.Context) error {
	log.Println("[SERVICE] Stopping queue...")

	svc.queueMutex.Lock()
	if !svc.queueClosed {
		close(svc.queue)
		svc.queueClosed = true
	}
	svc.queueMutex.Unlock()

	log.Println("[SERVICE] Waiting for workers to finish...")
	waitDone := make(chan struct{})
	go func() {
		svc.wg.Wait()
		close(waitDone)
	}()

	var pending []*linkTask
	select {
	case <-waitDone:
		log.Println("[SERVICE] Workers finished gracefully.")
	case <-ctx.Done():
		log.Printf("[SERVICE] Workers failed to finish before Fx deadline: %v", ctx.Err())
		log.Println("[SERVICE] Draining in-memory queue for persistence...")
		pending = svc.drainQueue()
	}

	svc.shutdownCancel()

	if len(pending) == 0 {
		log.Println("[SERVICE] No pending tasks to persist.")
		return nil
	}

	log.Println("[SERVICE] Saving task queue to disk...")
	if err := svc.SaveQueueToFile(pending); err != nil {
		return fmt.Errorf("failed to save task queue: %w", err)
	}

	log.Println("[SERVICE] Task queue saved successfully")
	return nil
}
