package opclients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/yaml.v3"
)

var lastAppliedConfigKey = "zk/last-applied-configuration"
var annotationsKey = "annotations"

func ApplyZerokObjects(path string) {
	yamlFiles, err := getAllYamlFileNamesInPath(path, true)

	if err != nil {
		panic("Error finding yaml files in given path")
	}
	createK8sObjects(path, yamlFiles)
}

func createK8sObjects(path string, fileNames []string) {
	for _, fn := range fileNames {
		yfile, err := ioutil.ReadFile(path + "/" + fn)

		if err != nil {

			log.Fatal(err)
		}

		yfilestring := string(yfile)

		parts := strings.Split(yfilestring, "\n---\n")

		for _, part := range parts {
			fmt.Println("Applying part", part)
			partBytes := []byte(part)
			yamlMap, err := yamlToMap(partBytes)
			if err != nil {
				fmt.Println("Caught error while converting yaml to map.")
				return
			}
			version, kind := getVK(yamlMap)
			namespace := getNamespace(yamlMap)
			name := getName(yamlMap)
			objectExist, respMap := doesObjectExist(version, kind, namespace, name)
			yamlMap = addLastAppliedConfiguration(yamlMap)
			if objectExist {
				fmt.Println("Object exists.")
				lastAppliedConfig := getLastAppliedConfig(respMap)
				jsonBytes, err := convertMapToJsonBytes(converMap(yamlMap))
				if err != nil {
					return
				}
				patch, err := findPatch([]byte(lastAppliedConfig), jsonBytes)
				if err != nil {
					fmt.Println("Error caught while finding patch for file ", fn, " error ", err)
				}
				updateObject(version, kind, namespace, name, patch)
			} else {
				fmt.Println("Object does not exist.")
				partBytes, err = yaml.Marshal(yamlMap)
				if err != nil {
					fmt.Println("Unable to marshal map into marshal after adding last applied config.")
					return
				}
				createObject(version, kind, namespace, partBytes)
			}
		}
	}
}

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

func getLastAppliedConfig(yamlMap map[string]interface{}) string {
	defaultOut := ""
	metadata, ok := yamlMap["metadata"]
	fmt.Printf("Metadata value is %v and type is %T\n", metadata, metadata)
	if ok {
		switch x := metadata.(type) {
		case map[string]interface{}:
			annotations, ok := x[annotationsKey]
			if ok {
				fmt.Printf("Annotations value is %v and type is %T\n", annotations, annotations)
				switch y := annotations.(type) {
				case map[string]interface{}:
					lastApplied, ok := y[lastAppliedConfigKey]
					if ok {
						switch z := lastApplied.(type) {
						case string:
							return z
						default:
							return defaultOut
						}
					} else {
						return defaultOut
					}
				default:
					return defaultOut
				}
			} else {
				return defaultOut
			}
		default:
			panic("metada of yaml object is not map.")
		}
	}
	return defaultOut
}

func addLastAppliedConfiguration(yamlMap map[interface{}]interface{}) map[interface{}]interface{} {
	metadata, ok := yamlMap["metadata"]
	if ok {
		switch x := metadata.(type) {
		case map[string]interface{}:
			annotations, ok := x["annotations"]
			if ok {
				switch y := annotations.(type) {
				case map[string]interface{}:
					yamlMapBytes, err := convertMapToJsonBytes(converMap(yamlMap))
					if err != nil {
						fmt.Println("Error caught while converting map to json.")
					}
					y[lastAppliedConfigKey] = string(yamlMapBytes)
					x[annotationsKey] = y
					yamlMap["metadata"] = x
					return yamlMap
				default:
					panic("Annotations in metadata of yaml object is not map.")
				}
			} else {
				yamlMapBytes, err := convertMapToJsonBytes(converMap(yamlMap))
				if err != nil {
					fmt.Println("Error caught while converting map to json.")
				}
				defaultOut := make(map[string]interface{})
				defaultOut[lastAppliedConfigKey] = string(yamlMapBytes)
				x[annotationsKey] = defaultOut
				yamlMap["metadata"] = x
				return yamlMap
			}
		default:
			panic("metada of yaml object is not map.")
		}
	}
	panic("metada is not present in yaml object.")
}

func getAllYamlFileNamesInPath(path string, recursive bool) ([]string, error) {
	files := []string{}

	items, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("Error caught while getting files from path ", err)
		return []string{}, fmt.Errorf("Error caught while getting files %v", err)
	}

	for _, f := range items {
		if f.IsDir() && recursive {
			subFiles, err := getAllYamlFileNamesInPath(path+"/"+f.Name(), recursive)
			if err != nil {
				fmt.Println("Error caught while getting files from path ", err)
				return []string{}, fmt.Errorf("Error caught while getting files %v", err)
			}
			files = append(files, subFiles...)
		} else if strings.HasSuffix(f.Name(), ".yaml") {
			files = append(files, f.Name())
		}
	}
	return files, nil
}

func findPatch(sourceJson []byte, targetJson []byte) ([]byte, error) {
	fmt.Printf("SourceJson %v.\n", string(sourceJson))
	fmt.Printf("targetJson %v.\n", string(targetJson))
	return jsonpatch.CreateMergePatch(sourceJson, targetJson)
}

// This method will get version and kind from a given map.
func getVK(yamlMap map[interface{}]interface{}) (string, string) {
	version, ok := yamlMap["apiVersion"]
	versionStr := ""
	if ok {
		switch x := version.(type) {
		case string:
			versionStr = x
		default:
			panic("Version of yaml object is not string.")
		}
	} else {
		panic("Version is not present in  yaml object.")
	}

	kind, ok := yamlMap["kind"]
	kindStr := ""
	if ok {
		switch x := kind.(type) {
		case string:
			kindStr = strings.ToLower(x) + "s"
		default:
			panic("Kind of yaml object is not string.")
		}
	} else {
		panic("Kind is not present in  yaml object.")
	}

	return versionStr, kindStr
}

func getName(yamlMap map[interface{}]interface{}) string {
	metadata, ok := yamlMap["metadata"]
	name := ""
	if ok {
		switch x := metadata.(type) {
		case map[string]interface{}:
			nameobj, ok := x["name"]
			if ok {
				switch y := nameobj.(type) {
				case string:
					name = y
				default:
					panic("name in yaml object is not string.")
				}
			}
		default:
			panic("metada of yaml object is not map.")
		}
	} else {
		panic("metadata is not present in  yaml object.")
	}
	return name
}

func getNamespace(yamlMap map[interface{}]interface{}) string {
	metadata, ok := yamlMap["metadata"]
	namespace := ""
	if ok {
		switch x := metadata.(type) {
		case map[string]interface{}:
			namespaceobj, ok := x["namespace"]
			if ok {
				switch y := namespaceobj.(type) {
				case string:
					namespace = y
				default:
					panic("namespace in yaml object is not string.")
				}
			}
		default:
			panic("metada of yaml object is not map.")
		}
	} else {
		panic("metadata is not present in  yaml object.")
	}
	return namespace
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
