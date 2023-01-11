package opclients

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/utils/env"
)

var token_path = "/var/run/secrets/kubernetes.io/serviceaccount/token"
var cert_path = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

type ApiServerClient struct {
	K8sServiceHost string
	K8sServicePort string
	caCertPath     string
	tokenPath      string
}

func doesObjectExist(version, kind, namespace string) bool {
	url := getUrl(version, kind, namespace)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Get request failed.")
	}

	req.Header.Add("Authorization", getBearerAuthToken())

	client := getHttpClient()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on response.\n[ERROR] -", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)

	code, ok := m["code"]

	if ok {
		switch x := code.(type) {
		case string:
			if x == "404" {
				return false
			}
		default:
			panic("Unkown response code received.")
		}
	}
	return true
}

func getUrl(version, kind, namespace string) string {
	url := ""
	if namespace == "" {
		url = "https://" + getK8sApiEndPoint() + "/apis/" + version + "/" + kind
	} else {
		url = "https://" + getK8sApiEndPoint() + "/apis/" + version + "/namespaces/" + namespace + "/" + kind
	}
	return url
}

func createObject(version, kind, namespace string, yaml []byte) {
	url := getUrl(version, kind, namespace)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yaml))

	if err != nil {
		fmt.Println("Creation request failed.")
	}

	req.Header.Add("Authorization", getBearerAuthToken())

	req.Header.Add("Content-Type", "application/yaml")

	client := getHttpClient()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on response.\n[ERROR] -", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	fmt.Println(string([]byte(body)))

}

func getHttpClient() *http.Client {
	caCertPool := getCaCertPool()
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
}

func getCaCertPool() *x509.CertPool {
	caCert, err := ioutil.ReadFile(cert_path)
	if err != nil {
		panic("Unable to ca cert file. Installation of zerok components failed.")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}

func getBearerAuthToken() string {
	tokenFile, err := ioutil.ReadFile(token_path)

	if err != nil {
		panic("Unable to find token file. Installation of zerok components failed.")
	}
	return "Bearer " + string(tokenFile)
}

func getK8sApiEndPoint() string {
	service_host := env.GetString("KUBERNETES_SERVICE_HOST", "")

	if service_host == "" {
		panic("Unable to find service host. Installation of zerok components failed.")
	}

	service_port := env.GetString("KUBERNETES_SERVICE_PORT", "")

	if service_port == "" {
		panic("Unable to find service port. Installation of zerok components failed.")
	}

	return service_host + ":" + service_port
}
