package auth

import (
	"fmt"
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	clientModel "github.com/zerok-ai/zk-utils-go/zkClient"
	"sync"
	"time"
)

var LOG_TAG = "ClusterTokenHandler"

var refreshTokenMutex sync.Mutex

type RefreshTokenCallback func()

var callbacks []RefreshTokenCallback

type ClusterTokenHandler struct {
	token           string
	clusterId       string
	zkConfig        config.ZkOperatorConfig
	killed          bool
	zkModules       []internal.ZkOperatorModule
	refreshingToken bool
}

func CreateClusterTokenHandler(config config.ZkOperatorConfig) *ClusterTokenHandler {
	opLogin := ClusterTokenHandler{}

	//Assigning initial values.
	opLogin.zkConfig = config
	opLogin.killed = false
	opLogin.token = ""
	opLogin.clusterId = ""
	opLogin.zkModules = make([]internal.ZkOperatorModule, 0)
	callbacks = []RefreshTokenCallback{}
	opLogin.refreshingToken = false
	return &opLogin
}

func (h *ClusterTokenHandler) GetClusterToken() string {
	return h.token
}

func (h *ClusterTokenHandler) GetClusterId() string {
	return h.clusterId
}

func (h *ClusterTokenHandler) isKilled() bool {
	return h.killed
}

func (h *ClusterTokenHandler) RefreshClusterToken(callback RefreshTokenCallback) error {
	logger.Info(LOG_TAG, "Refresh cluster token.")

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
		logger.Info(LOG_TAG, "Skipping refresh cluster token since cluster is killed.")
		return fmt.Errorf("cluster is killed")
	}
	clusterKey, err := utils.GetSecretValue(h.zkConfig.OperatorLogin.ClusterKeyNamespace, h.zkConfig.OperatorLogin.ClusterSecretName, h.zkConfig.OperatorLogin.ClusterKeyData)
	if err != nil {
		logger.Error(LOG_TAG, "Error while getting cluster key  from secrets :", err)
		return err
	}

	clusterTokenData, err := clientModel.DecodeToken(clusterKey)
	if err != nil {
		logger.Error(LOG_TAG, "Error while decoding cluster key data from secrets :", err)
		return err
	}

	h.token = clusterTokenData.TokenString

	killed, err := utils.GetSecretValue(h.zkConfig.OperatorLogin.ClusterKeyNamespace, h.zkConfig.OperatorLogin.ClusterSecretName, h.zkConfig.OperatorLogin.KilledData)
	if err != nil {
		logger.Error(LOG_TAG, "Error while getting killed key from secrets :", err)
		h.killed = false
	} else {
		h.killed = killed == "true"
	}

	clusterId, err := utils.GetSecretValue(h.zkConfig.OperatorLogin.ClusterKeyNamespace, h.zkConfig.OperatorLogin.ClusterSecretName, h.zkConfig.OperatorLogin.KilledData)
	if err != nil {
		logger.Error(LOG_TAG, "Error while getting cluster Id from secrets :", err)
	} else {
		h.clusterId = clusterId
	}

	refreshTokenMutex.Lock()
	defer refreshTokenMutex.Unlock()

	h.refreshingToken = false
	h.executeCallbackMethods()

	return nil
}

func (h *ClusterTokenHandler) executeCallbackMethods() {
	for _, callbackFunc := range callbacks {
		callbackFunc()
	}

	callbacks = make([]RefreshTokenCallback, 0)
}

func (h *ClusterTokenHandler) RegisterZkModules(modules []internal.ZkOperatorModule) {
	h.zkModules = modules
}

func (h *ClusterTokenHandler) deleteNamespaces(maxRetries int, retryDelay time.Duration) error {
	// TODO: commemting thi code temporarily.
	// Delete namespaces
	logger.Debug(LOG_TAG, "Deleting namespaces")
	//err := utils.DeleteNamespaceWithRetry("pl", maxRetries, retryDelay)
	//if err != nil {
	//	logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
	//	return err
	//}
	//err = utils.DeleteNamespaceWithRetry("px-operator", maxRetries, retryDelay)
	//if err != nil {
	//	logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
	//	return err
	//}
	//err = utils.DeleteNamespaceWithRetry("zk-client", maxRetries, retryDelay)
	//if err != nil {
	//	logger.Error(LOG_TAG, "Error while deleting namespace ", err.Error())
	//	return err
	//}
	return nil
}
