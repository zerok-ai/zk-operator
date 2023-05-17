package zkhttp

// TODO: Move to zkhttp
type ZkHttpError struct {
	Kind    string          `json:"kind,omitempty"`
	Message string          `json:"message,omitempty"`
	Param   string          `json:"param,omitempty"`
	Stack   any             `json:"stack,omitempty"`
	Info    *map[string]any `json:"info,omitempty"`
}

type ZkHttpResponse[T any] struct {
	Metadata *map[string]any    `json:"-"`
	Headers  *map[string]string `json:"-"`
	Status   int                `json:"-"`
	Error    *ZkHttpError       `json:"error,omitempty"`
	Message  *string            `json:"message,omitempty"`
	Debug    *map[string]any    `json:"debug,omitempty"`
	Data     T                  `json:"payload,omitempty"`
}

// Zk Http Response Builder
type ZkHttpResponseBuilder[T any] struct {
	_ZkHttpResponseBuilder _zkHttpResponseBuilder[T]
}

func (zkHttpResponseBuilder ZkHttpResponseBuilder[T]) WithStatus(status int) _zkHttpResponseBuilder[T] {
	zkHttpResponseBuilder._ZkHttpResponseBuilder = _zkHttpResponseBuilder[T]{}
	return zkHttpResponseBuilder._ZkHttpResponseBuilder.withStatus(status)
}

func (zkHttpResponseBuilder ZkHttpResponseBuilder[T]) WithZkErrorType(zkErrorType ZkErrorType) _zkHttpResponseBuilder[T] {
	zkHttpResponseBuilder._ZkHttpResponseBuilder = _zkHttpResponseBuilder[T]{}
	zkHttpError := ZkHttpError{}
	zkHttpError.Build(zkErrorType, nil, nil)
	return zkHttpResponseBuilder._ZkHttpResponseBuilder.withStatus(zkErrorType.Status).Error(zkHttpError)
}

// Zk Http Response Builder Internal
type _zkHttpResponseBuilder[T any] struct {
	ZkHttpResponse ZkHttpResponse[T]
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) withStatus(status int) _zkHttpResponseBuilder[T] {
	zkHttpResponseBuilder.ZkHttpResponse.Status = status
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Message(message *string) _zkHttpResponseBuilder[T] {
	zkHttpResponseBuilder.ZkHttpResponse.Message = message
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Data(data *T) _zkHttpResponseBuilder[T] {
	if data != nil {
		zkHttpResponseBuilder.ZkHttpResponse.Data = *data
	}
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Debug(key string, value any) _zkHttpResponseBuilder[T] {
	if !HTTP_DEBUG || value == nil {
		return zkHttpResponseBuilder
	}
	if zkHttpResponseBuilder.ZkHttpResponse.Debug == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Debug = &map[string]any{}
	}
	(*zkHttpResponseBuilder.ZkHttpResponse.Debug)[key] = value
	return zkHttpResponseBuilder
}

// func (zkHttpResponseBuilder *_zkHttpResponseBuilder) DebugP(key string, value any) *_zkHttpResponseBuilder{
// 	if !utils.HTTP_DEBUG {
// 		return zkHttpResponseBuilder;
// 	}
// 	if zkHttpResponseBuilder.ZkHttpResponse.Debug == nil {
// 		zkHttpResponseBuilder.ZkHttpResponse.Debug = &map[string]any{}
// 	}
// 	(*zkHttpResponseBuilder.ZkHttpResponse.Debug)[key] = value
// 	return zkHttpResponseBuilder;
// }

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Header(key string, value string) _zkHttpResponseBuilder[T] {
	if zkHttpResponseBuilder.ZkHttpResponse.Headers == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Headers = &map[string]string{}
	}
	(*zkHttpResponseBuilder.ZkHttpResponse.Headers)[key] = value
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Metadata(key string, value any) _zkHttpResponseBuilder[T] {
	if zkHttpResponseBuilder.ZkHttpResponse.Metadata == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Metadata = &map[string]any{}
	}
	(*zkHttpResponseBuilder.ZkHttpResponse.Metadata)[key] = value
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Error(zkHttpError ZkHttpError) _zkHttpResponseBuilder[T] {
	if zkHttpResponseBuilder.ZkHttpResponse.Error == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Error = &zkHttpError
	}
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) ErrorInfo(key string, value any) _zkHttpResponseBuilder[T] {
	if zkHttpResponseBuilder.ZkHttpResponse.Error == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Error = &ZkHttpError{}
	}
	if zkHttpResponseBuilder.ZkHttpResponse.Error.Info == nil {
		zkHttpResponseBuilder.ZkHttpResponse.Error.Info = &map[string]any{}
	}
	(*zkHttpResponseBuilder.ZkHttpResponse.Error.Info)[key] = value
	return zkHttpResponseBuilder
}

func (zkHttpResponseBuilder _zkHttpResponseBuilder[T]) Build() ZkHttpResponse[T] {
	return zkHttpResponseBuilder.ZkHttpResponse
}

func (zkHttpError *ZkHttpError) Build(zkErrorType ZkErrorType, param *string, stack any) {
	zkHttpError.Kind = zkErrorType.Type
	zkHttpError.Message = zkErrorType.Message
	if HTTP_DEBUG {
		zkHttpError.Stack = stack
	}
	if param != nil {
		zkHttpError.Param = *param
	}
}

func (zkHttpResponse ZkHttpResponse[T]) Header(key string, value string) {
	if zkHttpResponse.Headers == nil {
		zkHttpResponse.Headers = &map[string]string{}
	}
	(*zkHttpResponse.Headers)[key] = value
}

func (zkHttpResponse ZkHttpResponse[T]) IsOk() bool {
	if zkHttpResponse.Status > 199 && zkHttpResponse.Status < 300 {
		return true
	}
	return false
}
