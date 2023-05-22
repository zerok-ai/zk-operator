package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-utils-go/rules/model"
)

type SyncRules struct {
	redisClient *redis.Client
}

type RulesApiResponse struct {
	Payload FilterRulesObj `json:"payload"`
}

type FilterRulesObj struct {
	Rules   []model.FilterRule `json:"rules"`
	Deleted []string           `json:"deleted,omitempty"`
}

func (h *SyncRules) createNewRedisClient(config config.ZkInjectorConfig) {
	rulesConfig := config.RulesSync
	redisConfig := config.Redis
	addr := fmt.Sprint(redisConfig.Host, ":", redisConfig.Port)
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	fmt.Printf("Address for redis is %v.\n", addr)
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    "",
		DB:          rulesConfig.DB,
		ReadTimeout: readTimeout,
	})

	h.redisClient = _redisClient
}

func InitSyncRules(config config.ZkInjectorConfig) *SyncRules {
	syncRules := SyncRules{}
	syncRules.createNewRedisClient(config)
	return &syncRules
}

func (h *SyncRules) SyncRulesFromZkCloud(cfg config.ZkInjectorConfig) {
	h.getRulesFromZkCloud(cfg)
	//Creating a timer for periodic sync
	var duration = time.Duration(cfg.RulesSync.PollingInterval) * time.Second
	ticker := time.NewTicker(duration)
	for range ticker.C {
		fmt.Println("Sync rules triggered.")
		rules, err := h.getRulesFromZkCloud(cfg)
		if err != nil {
			fmt.Printf("Error while getting rules from zkcloud %v.\n", err)
			continue
		}
		err = h.saveRulesInRedis(rules)
		if err != nil {
			fmt.Printf("Error while savign rules to redis %v.\n", err)
		}
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

func (h *SyncRules) saveRulesInRedis(rulesApiResponse *RulesApiResponse) error {
	payload := rulesApiResponse.Payload
	for _, filterRule := range payload.Rules {
		fmt.Println("Printing filter rule.")
		fmt.Println(filterRule.FilterId)
		fmt.Println(filterRule.Workloads)
	}
	return nil
}
