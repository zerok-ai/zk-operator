package opclients

import structpb "github.com/golang/protobuf/ptypes/struct"

func GetRateLimiterValueStruct() *structpb.Struct {
	header := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"key": {
				Kind: &structpb.Value_StringValue{
					StringValue: "x-local-rate-limit",
				},
			},
			"value": {
				Kind: &structpb.Value_StringValue{
					StringValue: "true",
				},
			},
		},
	}
	response_headers := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"append": {
				Kind: &structpb.Value_BoolValue{
					BoolValue: false,
				},
			},
			"header": {
				Kind: &structpb.Value_StructValue{
					StructValue: header,
				},
			},
		},
	}
	defaultValue_FilterEnforced := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"numerator": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 100,
				},
			},
			"denominator": {
				Kind: &structpb.Value_StringValue{
					StringValue: "HUNDRED",
				},
			},
		},
	}
	filterEnforced := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"runtime_key": {
				Kind: &structpb.Value_StringValue{
					StringValue: "local_rate_limit_enforced",
				},
			},
			"default_value": {
				Kind: &structpb.Value_StructValue{
					StructValue: defaultValue_FilterEnforced,
				},
			},
		},
	}

	defaultValue_FilterEnabled := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"numerator": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 100,
				},
			},
			"denominator": {
				Kind: &structpb.Value_StringValue{
					StringValue: "HUNDRED",
				},
			},
		},
	}
	filterEnabled := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"runtime_key": {
				Kind: &structpb.Value_StringValue{
					StringValue: "local_rate_limit_enabled",
				},
			},
			"default_value": {
				Kind: &structpb.Value_StructValue{
					StructValue: defaultValue_FilterEnabled,
				},
			},
		},
	}
	tokenBucket := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"max_tokens": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 1,
				},
			},
			"tokens_per_fill": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 1,
				},
			},
			"fill_interval": {
				Kind: &structpb.Value_StringValue{
					StringValue: "2s",
				},
			},
		},
	}
	valueStructInternal := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"stat_prefix": {
				Kind: &structpb.Value_StringValue{
					StringValue: "http_local_rate_limiter",
				},
			},
			"token_bucker": {
				Kind: &structpb.Value_StructValue{
					StructValue: tokenBucket,
				},
			},
			"filter_enabled": {
				Kind: &structpb.Value_StructValue{
					StructValue: filterEnabled,
				},
			},
			"filter_enforced": {
				Kind: &structpb.Value_StructValue{
					StructValue: filterEnforced,
				},
			},
			"response_headers_to_add": {
				Kind: &structpb.Value_StructValue{
					StructValue: response_headers,
				},
			},
		},
	}
	typedConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"@type": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/udpa.type.v1.TypedStruct",
				},
			},
			"type_url": {
				Kind: &structpb.Value_StringValue{
					StringValue: "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit",
				},
			},
			"value": {
				Kind: &structpb.Value_StructValue{
					StructValue: valueStructInternal,
				},
			},
		},
	}
	valueStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": {
				Kind: &structpb.Value_StringValue{
					StringValue: "envoy.filters.http.local_ratelimit",
				},
			},
			"typed_config": {
				Kind: &structpb.Value_StructValue{
					StructValue: typedConfig,
				},
			},
		},
	}
	return valueStruct
}
