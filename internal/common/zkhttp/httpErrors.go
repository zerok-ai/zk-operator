package zkhttp

import (
	"net/http"

	"github.com/kataras/iris/v12"
)

type ZkErrorType struct {
	Message  string `json:"message"`
	Type     string `json:"type"`
	Status   int    `json:"status"`
	Metadata any    `json:"metadata"`
}

type ZkError struct {
	Error    ZkErrorType `json:"error"`
	Metadata any         `json:"metadata"`
}

// type _zkErrorBuilder struct {
// }

type ZkErrorBuilder struct {
}

var (
	ZK_ERROR_INTERNAL_SERVER = ZkErrorType{Status: iris.StatusInternalServerError,
		Type:    StatusText(iris.StatusInternalServerError),
		Message: "Encountered an issue, contact support"}
	ZK_ERROR_TIMEOUT = ZkErrorType{Status: iris.StatusRequestTimeout,
		Type:    StatusText(iris.StatusRequestTimeout),
		Message: "Encountered an issue, contact support"}
	ZK_ERROR_NOT_FOUND = ZkErrorType{Status: iris.StatusNotFound,
		Type:    StatusText(iris.StatusNotFound),
		Message: "Encountered an issue, contact support"}
	ZK_ERROR_SESSION_EXPIRED = ZkErrorType{Status: iris.StatusPageExpired,
		Type:    StatusText(iris.StatusPageExpired),
		Message: "The session has expired. Please login again to continue"}
	ZK_ERROR_UNAUTHORIZED = ZkErrorType{Status: iris.StatusUnauthorized,
		Type:    StatusText(iris.StatusUnauthorized),
		Message: "You are unauthorized to perform this operation. Contact system admin"}
	ZK_ERROR_BAD_REQUEST = ZkErrorType{Status: iris.StatusBadRequest,
		Type:    StatusText(iris.StatusBadRequest),
		Message: "The request was malformed or contained invalid parameters"}
	ZK_ERROR_INTERNAL_EMAIL_SERVER = ZkErrorType{Status: iris.StatusInternalServerError,
		Type:    StatusText(iris.StatusInternalServerError),
		Message: "Encountered an issue while sending email, contact support"}

	ZK_ERROR_DB_ERROR = ZkErrorType{Status: iris.StatusInternalServerError,
		Type:    StatusText(iris.StatusInternalServerError),
		Message: "Encountered an issue while executing db operation"}
)

func (zkError ZkError) SetMetadata(metadata any) {
	zkError.Metadata = metadata
}

var localStatusText = map[int]string{
	iris.StatusPageExpired: "Session Expired",
}

func StatusText(status int) string {
	statusText := http.StatusText(status)
	if statusText == "" {
		statusText = localStatusText[status]
	}

	return statusText
}

func (zkErrorBuilder ZkErrorBuilder) Build(zkErrorType ZkErrorType, metadata any) ZkError {
	return ZkError{
		Error:    zkErrorType,
		Metadata: metadata,
	}
}

func (zkErrorBuilder ZkErrorBuilder) CreateZkErrorType(status int, message string) ZkErrorType {
	return ZkErrorType{
		Status:  status,
		Type:    StatusText(status),
		Message: message,
	}
}
