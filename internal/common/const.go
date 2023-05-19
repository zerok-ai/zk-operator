package common

const (
	ZkOrchKey          = "zk-status"
	ZkOrchPath         = "/metadata/labels/" + ZkOrchKey
	ZkOrchOrchestrated = "orchestrated"
	ZkOrchProcessed    = "processed"
	ZkOrchInProcess    = "in-process"

	JavalToolOptions = "JAVA_TOOL_OPTIONS"
	OtelArgument     = " -javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar"

	ZkInjectionKey   = "zk-injection"
	ZkInjectionValue = "enabled"

	ZkAutoRestartKey = "zk-auto-restart"

	ZkConfigMapName = "zk-image-configmap"
	ZkConfigMapKey  = "zk-image-map"

	NamespaceEnvVariable = "POD_NAMESPACE"

	HashSetName       string = "zk_img_proc_map"
	HashSetVersionKey string = "zk_img_proc_version"
)
