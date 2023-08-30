package common

import "time"

const (
	ZkOrchKey          = "zk-status"
	ZkOrchPath         = "/metadata/labels/" + ZkOrchKey
	ZkOrchOrchestrated = "orchestrated"
	ZkOrchProcessed    = "processed"
	ZkOrchInProcess    = "in-process"

	JavalToolOptions = "JAVA_TOOL_OPTIONS"
	ZkInjectionKey   = "zk-injection"
	ZkInjectionValue = "enabled"

	ZkAutoRestartKey   = "zk-auto-restart"
	ZkAutoRestartValue = "enabled"

	ZkImageConfigMapName = "zk-image"
	ZkImageConfigMapKey  = "zk-image-map"

	ZkProcessConfigMapName = "zk-process-info"
	ZkProcessConfigMapKey  = "processes"

	NamespaceEnvVariable = "POD_NAMESPACE"

	HashSetName       string = "zk_img_proc_map"
	HashSetVersionKey string = "zk_img_proc_version"

	OperatorTokenHeaderKey string = "Operator-Auth-Token"

	//Kill switch configuration
	NamespaceDeleteRetryLimit = 3
	NamespaceDeleteRetryDelay = 2 * time.Second

	ScenarioSyncInterval = 5 * time.Minute

	RedisImageDbName   = "image_db"
	RedisVersionDbName = "version_db"
)
