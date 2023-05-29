package sync

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-utils-go/rules/model"
)

var authTokenExpiredCode = 401

type SyncRules struct {
	VersionedStore *storage.VersionedStore
	OpLogin        *auth.OperatorLogin
	ticker         *time.Ticker
}

type RulesApiResponse struct {
	Payload ScenariosObj `json:"payload"`
}

type ScenariosObj struct {
	Scenarios []model.Scenario `json:"scenarios"`
	Deleted   []string         `json:"deleted,omitempty"`
}

func (h *SyncRules) Init(VersionedStore *storage.VersionedStore, OpLogin *auth.OperatorLogin) {
	h.VersionedStore = VersionedStore
	h.OpLogin = OpLogin
	//h.killed = false
}

func (h *SyncRules) SyncRulesFromZkCloud(cfg config.ZkInjectorConfig) {
	h.updateRules(cfg)
	//Creating a timer for periodic sync
	var duration = time.Duration(cfg.RulesSync.PollingInterval) * time.Second
	h.ticker = time.NewTicker(duration)
	for range h.ticker.C {
		fmt.Println("Sync rules triggered.")
		h.updateRules(cfg)
	}
}

func (h *SyncRules) updateRules(cfg config.ZkInjectorConfig) {
	rules, err := h.getRulesFromZkCloud(cfg)
	if err != nil {
		fmt.Printf("Error while getting rules from zkcloud %v.\n", err)
		return
	}
	err = h.updateRulesInRedis(rules)
	if err != nil {
		fmt.Printf("Error while savign rules to redis %v.\n", err)
	}
}

func (h *SyncRules) getRulesFromZkCloud(cfg config.ZkInjectorConfig) (*RulesApiResponse, error) {

	fmt.Println("Get rules from zk cloud.")

	endpoint := "http://" + cfg.RulesSync.Host + cfg.RulesSync.Path
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	fmt.Printf("Current operator token is %v.\n", h.OpLogin.GetOperatorToken())

	if h.OpLogin.GetOperatorToken() == "" {
		return nil, h.refreshAuthToken(cfg)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request for rules api :", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		return nil, h.refreshAuthToken(cfg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response from rules api :", err)
		return nil, err
	}

	var apiResponse RulesApiResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("Error while unmarshalling rules api response :", err)
		return nil, err
	}

	return &apiResponse, nil
}

func (h *SyncRules) refreshAuthToken(cfg config.ZkInjectorConfig) error {
	err := h.OpLogin.RefreshOperatorToken(func() {
		h.updateRules(cfg)
	})
	if err != nil {
		fmt.Printf("Error while refreshing auth token %v.\n", err)
	}
	return err
}

func (h *SyncRules) updateRulesInRedis(rulesApiResponse *RulesApiResponse) error {
	payload := rulesApiResponse.Payload
	latestVersion := "0"
	for _, scenario := range payload.Scenarios {
		ver1, err1 := strconv.ParseInt(latestVersion, 10, 64)
		ver2, err2 := strconv.ParseInt(scenario.Version, 10, 64)
		if err1 != nil || err2 != nil {
			fmt.Printf("Error while converting versions to int64 for scenario %v.\n", scenario.ScenarioId)
			continue
		}

		if ver2 > ver1 {
			latestVersion = scenario.Version
		}

		scenarioString, err := json.Marshal(scenario)
		if err != nil {
			fmt.Printf("Error while converting filter rule to string %v.\n", err)
			return err
		}
		scenarioId := scenario.ScenarioId
		err = h.VersionedStore.SetValue(scenarioId, string(scenarioString))
		if err != nil {
			fmt.Printf("Error while setting filter rule to redis %v.\n", err)
			return err
		}
	}

	for _, scenarioId := range payload.Deleted {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			fmt.Printf("Error while deleting filter id %v from redis %v.\n", scenarioId, err)
			return err
		}
	}

	return nil
}

func (h *SyncRules) CleanUpOnkill() error {
	fmt.Printf("Kill method in sync rules.\n")
	h.ticker.Stop()
	return nil
}
