package opclients

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

func converMap(yamlMap map[interface{}]interface{}) map[string]interface{} {
	outMap := make(map[string]interface{})
	for k, v := range yamlMap {
		switch x := k.(type) {
		case string:
			outMap[x] = v
		default:
		}
	}

	return outMap
}

func convertMapToJsonBytes(yamlMap map[string]interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(yamlMap)
	if err != nil {
		fmt.Println("Unable to convert map to json.")
		return nil, err
	}
	return jsonBytes, nil
}

func getMetadata(yamlMap map[interface{}]interface{}) (map[string]interface{}, error) {
	metadata, ok := yamlMap["metadata"]
	notFoundError := fmt.Errorf("no metadata element found")
	if ok {
		switch x := metadata.(type) {
		case map[string]interface{}:
			return x, nil
		default:
			return nil, notFoundError
		}
	}
	return nil, notFoundError
}

func yamlToMap(data []byte) (map[interface{}]interface{}, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		fmt.Printf("error while converting yaml: %v.\n", err)
		return m, err
	}
	return m, nil
}
