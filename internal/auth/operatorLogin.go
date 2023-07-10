package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"io"
	"net/http"
	"reflect"
	"strconv"
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
	zkConfig         config.ZkOperatorConfig
	killed           bool
	refreshingToken  bool
	zkModules        []internal.ZkOperatorModule
	lastTokenRefresh time.Time
}

type OperatorLoginResponse struct {
	Payload OperatorTokenObj    `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

type OperatorTokenObj struct {
	Token     string `json:"operatorAuthToken"`
	ClusterId string `json:"clusterId"`
	Killed    bool   `json:"killed"`
}

type OperatorLoginRequest struct {
	ClusterKey string `json:"clusterKey"`
}

func CreateOperatorLogin(config config.ZkOperatorConfig) *OperatorLogin {
	opLogin := OperatorLogin{}

	//Assigning initial values.
	opLogin.zkConfig = config
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

	maxRetries := h.zkConfig.OperatorLogin.MaxRetries
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
	port := h.zkConfig.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}
	endpoint := protocol + "://" + h.zkConfig.ZkCloud.Host + ":" + h.zkConfig.ZkCloud.Port + h.zkConfig.OperatorLogin.Path

	clusterKey, err := utils.GetSecretValue(h.zkConfig.OperatorLogin.ClusterKeyNamespace, h.zkConfig.OperatorLogin.ClusterSecretName, h.zkConfig.OperatorLogin.ClusterKeyData)

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

	if !utils.RespCodeIsOk(resp.StatusCode) {
		message := "response code is not ok for operator login api - " + strconv.Itoa(resp.StatusCode)
		logger.Error(LOG_TAG, message)
		return fmt.Errorf(message)
	}

	var apiResponse OperatorLoginResponse
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		logger.Error(LOG_TAG, "Error while unmarshalling rules operator login api response :", err)
		return err
	}

	if apiResponse.Error != nil {
		message := "found error in operator login api response " + apiResponse.Error.Message
		logger.Error(LOG_TAG, message)
		return fmt.Errorf(message)
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

	logger.Debug(LOG_TAG, "ClusterId is ", h.clusterId)
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
