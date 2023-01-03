package opclients

import structpb "github.com/golang/protobuf/ptypes/struct"

func GetLogValueStruct() *structpb.Struct {
	accessFile := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"@type": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
				},
			},
			"path": {
				Kind: &structpb.Value_StringValue{
					StringValue: "/dev/stdout",
				},
			},
			"format": {
				Kind: &structpb.Value_StringValue{
					StringValue: "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \n",
				},
			},
		},
	}
	acessLog := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": {
				Kind: &structpb.Value_StringValue{
					StringValue: "envoy.access_loggers.file",
				},
			},
			"typed_config": {
				Kind: &structpb.Value_StructValue{
					StructValue: accessFile,
				},
			},
		},
	}
	acessLogValue := &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: acessLog,
		},
	}
	typedConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"@type": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
				},
			},
			"access_log": {
				Kind: &structpb.Value_ListValue{
					ListValue: &structpb.ListValue{
						Values: []*structpb.Value{acessLogValue},
					},
				},
			},
		},
	}
	valueStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"typed_config": {
				Kind: &structpb.Value_StructValue{
					StructValue: typedConfig,
				},
			},
		},
	}
	return valueStruct
}
