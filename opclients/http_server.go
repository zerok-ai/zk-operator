package opclients

import (
	"fmt"
	"io"
	"net/http"
)

func StartServer() {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Exception request received.")
		requestBody, _ := io.ReadAll(r.Body)
		fmt.Println("Request body is ", string(requestBody))
		w.WriteHeader(200)
		w.Write([]byte("Done"))
	})
	http.Handle("/exception", h)
	http.ListenAndServe(":8127", nil)
}
