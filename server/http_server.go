package server

import "net/http"

func startServer() {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Done"))
	})
	http.Handle("/exception", h)
	http.ListenAndServe(":8127", nil)
}
