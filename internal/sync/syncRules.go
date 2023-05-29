package sync

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"io"
	"net/http"
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
	Payload FilterRulesObj `json:"payload"`
}

type FilterRulesObj struct {
	Rules   []model.FilterRule `json:"rules"`
	Deleted []string           `json:"deleted,omitempty"`
}

func CreateSyncRules(VersionedStore *storage.VersionedStore, OpLogin *auth.OperatorLogin) *SyncRules {
	syncRules := SyncRules{}
	syncRules.VersionedStore = VersionedStore
	syncRules.OpLogin = OpLogin
	return &syncRules
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

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request for rules api :", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		err := h.OpLogin.RefreshOperatorToken(func() {
			h.updateRules(cfg)
		})
		if err != nil {
			fmt.Printf("Error while refreshing auth token %v.\n", err)
			return nil, err
		}
		return nil, fmt.Errorf("Auth token expired %v.\n", err)
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

func (h *SyncRules) updateRulesInRedis(rulesApiResponse *RulesApiResponse) error {
	payload := rulesApiResponse.Payload
	for _, filterRule := range payload.Rules {
		filterString, err := json.Marshal(filterRule)
		if err != nil {
			fmt.Printf("Error while converting filter rule to string %v.\n", err)
			return err
		}
		filterId := filterRule.FilterId
		err = h.VersionedStore.SetValue(filterId, string(filterString))
		if err != nil {
			fmt.Printf("Error while setting filter rule to redis %v.\n", err)
			return err
		}
	}

	for _, filterId := range payload.Deleted {
		err := h.VersionedStore.Delete(filterId)
		if err != nil {
			fmt.Printf("Error while deleting filter id %v from redis %v.\n", filterId, err)
			return err
		}
	}

	return nil
}

func (h *SyncRules) CleanUpOnkill() error {
	h.ticker.Stop()
	return nil
}
