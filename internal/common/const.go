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

	ZkAutoRestartKey = "zk-auto-restart"

	ZkConfigMapName = "zk-image-configmap"
	ZkConfigMapKey  = "zk-image-map"

	NamespaceEnvVariable = "POD_NAMESPACE"

	HashSetName       string = "zk_img_proc_map"
	HashSetVersionKey string = "zk_img_proc_version"

	OperatorTokenHeaderKey string = "Operator-Auth-Token"

	//Kill switch configuration
	NamespaceDeleteRetryLimit = 3
	NamespaceDeleteRetryDelay = 2 * time.Second

	RedisImageDbName   = "image_db"
	RedisVersionDbName = "version_db"
)
