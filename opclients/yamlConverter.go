package opclients

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/yaml.v3"
)

// TODO: get this path from yaml file of operator
var path = ""

func createK8sObjects(fileNames []string) {
	for _, fn := range fileNames {
		yfile, err := ioutil.ReadFile(path + "/" + fn)

		if err != nil {

			log.Fatal(err)
		}

		yfilestring := string(yfile)

		parts := strings.Split(yfilestring, "\n---\n")

		for _, part := range parts {
			partBytes := []byte(part)
			yamlMap, err := yamlToMap(partBytes)
			if err != nil {
				fmt.Println("Caught error while converting yaml to map.")
				return
			}
			version, kind := getVK(yamlMap)
			namespace := getNamespace(yamlMap)
			if doesObjectExist(version, kind, namespace) {
				//get last applied configuration
				// create patch
				// update last applied configuration value in patch
				// apply patch
			} else {
				// add last applied configuration in patch
				yamlMap = addLastAppliedConfiguration(yamlMap)
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

func getExistingAnnotationsMap(yamlMap map[interface{}]interface{}) map[interface{}]interface{} {
	defaultOut := make(map[interface{}]interface{})
	metadata, ok := yamlMap["metadata"]
	if ok {
		switch x := metadata.(type) {
		case map[interface{}]interface{}:
			annotations, ok := x["annotations"]
			if ok {
				switch y := annotations.(type) {
				case map[interface{}]interface{}:
					return y
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
		case map[interface{}]interface{}:
			annotations, ok := x["annotations"]
			if ok {
				switch y := annotations.(type) {
				case map[interface{}]interface{}:
					y["zk/last-applied-configuration"] = fmt.Sprint(yamlMap)
					x["annotatations"] = y
					return x
				default:
					panic("Annotations in metadata of yaml object is not map.")
				}
			} else {
				defaultOut := make(map[interface{}]interface{})
				defaultOut["zk/last-applied-configuration"] = fmt.Sprint(yamlMap)
				x["annotatations"] = defaultOut
				return x
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

func getNamespace(yamlMap map[interface{}]interface{}) string {
	metadata, ok := yamlMap["metadata"]
	namespace := ""
	if ok {
		switch x := metadata.(type) {
		case map[interface{}]interface{}:
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
