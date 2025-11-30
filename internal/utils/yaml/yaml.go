package yaml

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	//"userbot/internal/utils/files"
	"link-availability-checker/internal/utils/closer"
	"link-availability-checker/internal/utils/files"
)

// GetBool returns a boolean value from a YAML file by key
//
// Used in config.MuteFxLog() for disabling Fx logs before configuration is loaded
func GetBool(filePath string, key string) bool {
	value := getKey(filePath, key)
	boolValue, _ := value.(bool) // No comma-ok idiom to return false if type assert failed
	return boolValue
}

func getKey(filePath string, key string) interface{} {
	if !files.FileExists(filePath) {
		return false
	}
	if files.FileIsEmpty(filePath) {
		return false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer closer.Close(file)

	decoder := yaml.NewDecoder(file)
	var config map[string]interface{}
	if err = decoder.Decode(&config); err != nil {
		return false
	}

	keys := strings.Split(key, ".")
	var value interface{} = config
	for _, k := range keys {
		if subMap, ok := value.(map[string]interface{}); ok {
			value = subMap[k]
		} else {
			return false
		}
	}

	return value
}
