package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"
)

var refreshTokenMutex sync.Mutex

type RefreshTokenCallback func()

var callbacks []RefreshTokenCallback

type OperatorLogin struct {
	operatorToken   string
	operatorConfig  config.OperatorLoginConfig
	killed          bool
	refreshingToken bool
	zkModules       []internal.ZkModule
}

type OperatorLoginResponse struct {
	Payload OperatorTokenObj `json:"payload"`
}

type OperatorTokenObj struct {
	Token  string `json:"operatorAuthToken"`
	Killed bool   `json:"killed"`
}

type OperatorLoginRequest struct {
	ClusterKey string `json:"clusterKey"`
}

func CreateOperatorLogin(config config.OperatorLoginConfig) *OperatorLogin {
	opLogin := OperatorLogin{operatorConfig: config, killed: false, operatorToken: "", zkModules: []internal.ZkModule{}}
	callbacks = []RefreshTokenCallback{}
	return &opLogin
}

func (h *OperatorLogin) GetOperatorToken() string {
	return h.operatorToken
}

func (h *OperatorLogin) isKilled() bool {
	return h.killed
}

func (h *OperatorLogin) RefreshOperatorToken(callback RefreshTokenCallback) error {
	fmt.Println("Request operator token.")
	refreshTokenMutex.Lock()

	callbacks = append(callbacks, callback)

	if h.refreshingToken {
		// Another refresh token request is already in progress.
		return nil
	}

	h.refreshingToken = true

	refreshTokenMutex.Unlock()

	if h.killed {
		fmt.Println("Skipping refresh access token api since cluster is killed.")
		return nil
	}

	endpoint := "http://" + h.operatorConfig.Host + h.operatorConfig.Path

	clusterKey, err := utils.GetSecretValue(h.operatorConfig.ClusterKeyNamespace, h.operatorConfig.ClusterKey, h.operatorConfig.ClusterKeyData)

	if err != nil {
		fmt.Println("Error while getting cluster key from secrets :", err)
		return err
	}

	requestPayload := OperatorLoginRequest{ClusterKey: clusterKey}

	data, err := json.Marshal(requestPayload)

	fmt.Println("Request payload ", string(data))

	if err != nil {
		fmt.Println("Error while creating payload for operator login request:", err)
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))

	if err != nil {
		fmt.Println("Error creating operator login request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error sending request for operator login api :", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response from operator login api :", err)
		return err
	}

	fmt.Println("Operator response ", string(body))

	var apiResponse OperatorLoginResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("Error while unmarshalling rules operator login api response :", err)
		return err
	}

	if apiResponse.Payload.Killed {
		fmt.Println("Api response came as killed.")
		h.killed = true
		for _, module := range h.zkModules {
			err := module.CleanUpOnkill()
			if err != nil {
				fmt.Printf("Error while cleaning up on kill method for module %v.\n", reflect.TypeOf(module).Name())
			}
		}
		return h.deleteNamespaces(common.NamespaceDeleteRetryLimit, common.NamespaceDeleteRetryDelay)
	}

	h.operatorToken = apiResponse.Payload.Token

	fmt.Println("Token is ", h.operatorToken)

	refreshTokenMutex.Lock()

	h.refreshingToken = false

	for _, callbackFunc := range callbacks {
		callbackFunc()
	}

	callbacks = make([]RefreshTokenCallback, 0)

	refreshTokenMutex.Unlock()

	return nil
}

func (h *OperatorLogin) RegisterZkModules(modules []internal.ZkModule) {
	h.zkModules = modules
}

func (h *OperatorLogin) deleteNamespaces(maxRetries int, retryDelay time.Duration) error {
	err := utils.DeleteNamespaceWithRetry("pl", maxRetries, retryDelay)
	if err != nil {
		fmt.Printf("Error while deleting namespace %v \n", err)
		return err
	}
	err = utils.DeleteNamespaceWithRetry("px-operator", maxRetries, retryDelay)
	if err != nil {
		fmt.Printf("Error while deleting namespace %v \n", err)
		return err
	}
	err = utils.DeleteNamespaceWithRetry("zk-client", maxRetries, retryDelay)
	if err != nil {
		fmt.Printf("Error while deleting namespace %v \n", err)
		return err
	}
	return nil
}
