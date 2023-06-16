package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"
)

var LOG_TAG = "OperatorLogin"

var refreshTokenMutex sync.Mutex

type RefreshTokenCallback func()

var callbacks []RefreshTokenCallback

type OperatorLogin struct {
	operatorToken    string
	clusterId        string
	operatorConfig   config.OperatorLoginConfig
	killed           bool
	refreshingToken  bool
	zkModules        []internal.ZkOperatorModule
	lastTokenRefresh time.Time
}

type OperatorLoginResponse struct {
	Payload OperatorTokenObj `json:"payload"`
}

type OperatorTokenObj struct {
	Token     string `json:"operatorAuthToken"`
	ClusterId string `json:"clusterId"`
	Killed    bool   `json:"killed"`
}

type OperatorLoginRequest struct {
	ClusterKey string `json:"clusterKey"`
}

func CreateOperatorLogin(config config.OperatorLoginConfig) *OperatorLogin {
	opLogin := OperatorLogin{}

	//Assigning initial values.
	opLogin.operatorConfig = config
	opLogin.killed = false
	opLogin.operatorToken = ""
	opLogin.zkModules = []internal.ZkOperatorModule{}
	opLogin.clusterId = ""
	callbacks = []RefreshTokenCallback{}

	return &opLogin
}

func (h *OperatorLogin) GetOperatorToken() string {
	return h.operatorToken
}

func (h *OperatorLogin) GetClusterId() string {
	return h.clusterId
}

func (h *OperatorLogin) isKilled() bool {
	return h.killed
}

func (h *OperatorLogin) RefreshOperatorToken(callback RefreshTokenCallback) error {
	logger.Info(LOG_TAG, "Request operator token.")
	refreshTokenMutex.Lock()

	callbacks = append(callbacks, callback)

	if h.refreshingToken {
		refreshTokenMutex.Unlock()
		// Another refresh token request is already in progress.
		return fmt.Errorf("another refresh token request is already in progress")
	}

	h.refreshingToken = true

	refreshTokenMutex.Unlock()

	if h.killed {
		logger.Info(LOG_TAG, "Skipping refresh access token api since cluster is killed.")
		return fmt.Errorf("cluster is killed")
	}

	maxRetries := h.operatorConfig.MaxRetries
	retryCount := 0

	for retryCount <= maxRetries {
		err2 := h.getOpTokenFromZkCloud()
		if err2 != nil {
			retryCount++
		} else {
			break
		}
	}

	refreshTokenMutex.Lock()
	defer refreshTokenMutex.Unlock()

	h.refreshingToken = false
	h.executeCallbackMethods()

	return nil
}

func (h *OperatorLogin) executeCallbackMethods() {
	for _, callbackFunc := range callbacks {
		callbackFunc()
	}

	callbacks = make([]RefreshTokenCallback, 0)
}

func (h *OperatorLogin) getOpTokenFromZkCloud() error {
	endpoint := "http://" + h.operatorConfig.Host + h.operatorConfig.Path

	clusterKey, err := utils.GetSecretValue(h.operatorConfig.ClusterKeyNamespace, h.operatorConfig.ClusterKey, h.operatorConfig.ClusterKeyData)

	if err != nil {
		logger.Error(LOG_TAG, "Error while getting cluster key from secrets :", err)
		return err
	}

	requestPayload := OperatorLoginRequest{ClusterKey: clusterKey}

	data, err := json.Marshal(requestPayload)

	if err != nil {
		logger.Error(LOG_TAG, "Error while creating payload for operator login request:", err)
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))

	if err != nil {
		logger.Error(LOG_TAG, "Error creating operator login request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		logger.Error(LOG_TAG, "Error sending request for operator login api :", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.Error(LOG_TAG, "Error reading response from operator login api :", err)
		return err
	}

	var apiResponse OperatorLoginResponse
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		logger.Error(LOG_TAG, "Error while unmarshalling rules operator login api response :", err)
		return err
	}

	if apiResponse.Payload.Killed {
		logger.Info(LOG_TAG, "Api response came as killed.")
		h.killed = true
		for _, module := range h.zkModules {
			err := module.CleanUpOnkill()
			if err != nil {
				logger.Error(LOG_TAG, "Error while cleaning up on kill method for module ", reflect.TypeOf(module).Name())
			}
		}
		return h.deleteNamespaces(common.NamespaceDeleteRetryLimit, common.NamespaceDeleteRetryDelay)
	}

	h.operatorToken = apiResponse.Payload.Token
	h.clusterId = apiResponse.Payload.ClusterId
	return nil
}

func (h *OperatorLogin) RegisterZkModules(modules []internal.ZkOperatorModule) {
	h.zkModules = modules
}

func (h *OperatorLogin) deleteNamespaces(maxRetries int, retryDelay time.Duration) error {
	err := utils.DeleteNamespaceWithRetry("pl", maxRetries, retryDelay)
	if err != nil {
		logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
		return err
	}
	err = utils.DeleteNamespaceWithRetry("px-operator", maxRetries, retryDelay)
	if err != nil {
		logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
		return err
	}
	err = utils.DeleteNamespaceWithRetry("zk-client", maxRetries, retryDelay)
	if err != nil {
		logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
		return err
	}
	return nil
}
