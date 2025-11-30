package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"

	"link-availability-checker/internal/utils/yaml"
)

const DefaultConfigLocation = "./config.yaml"

func LoadConfig() {
	viper.SetConfigFile(DefaultConfigLocation)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}
	if err := ValidateConfigFields(); err != nil {
		log.Fatalf("Failed to validate config: %s", err)
	}
}

const (
	LogFilePath  = "app.log.path"           // string
	MuteFx       = "app.log.mute_fx"        // bool
	MuteGinDebug = "app.log.mute_gin_debug" // bool

	LinksFilePath = "app.filestore.path" // string

	ApiPort     = "app.api.port"      // int
	ApiBasePath = "app.api.base_path" // string
	ApiPassword = "app.api.password"  // string

	QueueFilePath = "app.queue.path"    // string
	QueueWorkers  = "app.queue.workers" // int

	WorkersRatio = "app.worker_pool.workers_ratio" // int
	MaxWorkers   = "app.worker_pool.workers_limit" // int

	RecheckStatusesWhenPrinting = "app.links.recheck_statuses_on_print" // bool
)

func ValidateConfigFields() error {
	required := []string{ApiPort, LogFilePath, LinksFilePath, QueueFilePath, QueueWorkers, WorkersRatio, MaxWorkers}
	var missing []string

	for _, key := range required {
		if viper.GetString(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		if len(missing) == 1 {
			return fmt.Errorf("missing or empty config field: %s", missing[0])
		} else {
			return fmt.Errorf("missing or empty config fields: %s", strings.Join(missing, ", "))
		}
	}

	if viper.GetInt(WorkersRatio) == 0 {
		return fmt.Errorf("key \"%s\" must not be 0", WorkersRatio)
	} // Division by zero prevention

	return nil
}

func MuteFxLog() fx.Option {
	if yaml.GetBool(DefaultConfigLocation, MuteFx) {
		return fx.Options(fx.NopLogger)
	} else {
		return fx.Options()
	}
}
