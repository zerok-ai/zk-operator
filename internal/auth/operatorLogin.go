package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"io"
	"net/http"
	"sync"
)

var refreshTokenMutex sync.Mutex

type RefreshTokenCallback interface {
	RefreshTokenCallback()
}

type RefreshTokenCallbackFunc func()

var callbackFuncs []RefreshTokenCallbackFunc

type OperatorLogin struct {
	operatorToken   string
	operatorConfig  config.OperatorLoginConfig
	killed          bool
	refreshingToken bool
	zkmodules       []internal.Zkmodule
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
	opLogin := OperatorLogin{operatorConfig: config, killed: false, operatorToken: ""}
	callbackFuncs = []RefreshTokenCallbackFunc{}
	return &opLogin
}

func (h *OperatorLogin) GetOperatorToken() string {
	return h.operatorToken
}

func (h *OperatorLogin) isKilled() bool {
	return h.killed
}

func (h *OperatorLogin) RefreshOperatorToken(callback RefreshTokenCallbackFunc) error {

	refreshTokenMutex.Lock()

	callbackFuncs = append(callbackFuncs, callback)

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
		h.killed = true
		for _, module := range h.zkmodules {
			module.CleanUpOnkill()
		}
		return nil
	}

	h.operatorToken = apiResponse.Payload.Token

	fmt.Println("Token is ", h.operatorToken)

	refreshTokenMutex.Lock()

	h.refreshingToken = false

	for _, callbackFunc := range callbackFuncs {
		callbackFunc()
	}

	callbackFuncs = []RefreshTokenCallbackFunc{}

	refreshTokenMutex.Unlock()

	return nil
}

func (h *OperatorLogin) registerZkModules(modules []internal.Zkmodule) {
	h.zkmodules = modules
}
