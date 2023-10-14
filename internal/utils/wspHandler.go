package utils

import (
	"github.com/zerok-ai/zk-operator/internal/config"
	"net/http"
	"net/url"
)

func RouteRequestFromWspClient(request *http.Request, config config.ZkOperatorConfig) (*http.Response, error) {
	wspConfig := config.WspClient

	fullDestinationUrl := getFullUrl(request)
	newURL := "http://" + wspConfig.Host + ":" + wspConfig.Port + wspConfig.Path

	request.URL, _ = request.URL.Parse(newURL)
	request.Header.Add(wspConfig.DestinationHeader, fullDestinationUrl)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getFullUrl(r *http.Request) string {

	reqURL := r.URL
	queryParams := reqURL.Query()

	u := url.URL{
		Scheme:   reqURL.Scheme,
		Host:     reqURL.Host,
		Path:     reqURL.Path,
		RawQuery: queryParams.Encode(),
	}
	return u.String()
}
