package signals

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func SendInterruptSignal() {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Printf("Failed to find own process: %v", err)
		return
	}

	err = p.Signal(os.Interrupt)
	if err != nil {
		log.Printf("Failed to send INTERRUPT signal to own proccess: %v", err)
		return
	}
}

func RestartSelf() error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get own executable path: %w", err)
	}

	args := ""
	for _, a := range os.Args[1:] {
		args += fmt.Sprintf(" %s", a)
	}

	const delaySeconds = 10
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sleep %d && nohup %s%s &", delaySeconds, self, args))

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}
	return nil
}
