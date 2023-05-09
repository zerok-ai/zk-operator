package common

const (
	ZkOrchKey          = "zk-status"
	ZkOrchPath         = "/metadata/labels/" + ZkOrchKey
	ZkOrchOrchestrated = "orchestrated"
	ZkOrchProcessed    = "processed"
	ZkOrchInProcess    = "in-process"

	JavalToolOptions = "JAVA_TOOL_OPTIONS"
	OtelArgument     = " -javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar -Dotel.traces.exporter=jaeger -Dotel.exporter.jaeger.endpoint=http://simplest-collector.observability.svc.cluster.local:14250"

	ZkInjectionKey   = "zk-injection"
	ZkInjectionValue = "enabled"

	ZkAutoRestartKey = "zk-auto-restart"
)
