package resources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var port = ":9093"

func RegisterUpdateEnvoyAPI() {
	router := mux.NewRouter()

	router.HandleFunc("/updenvoy/", UpdateEnvoyConfig).Methods("POST")

	fmt.Printf("Server at %v.\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func UpdateEnvoyConfig(w http.ResponseWriter, r *http.Request) {
	var updateEnvoyConfigRequest UpdateEnvoyConfRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Error while reading request body in update envoy request.")
		createErrorResponse(w, "", "Error while reading request body.")
	}
	json.Unmarshal(reqBody, &updateEnvoyConfigRequest)
	createSuccessResponse(w)
}

func createErrorResponse(w http.ResponseWriter, errorType string, errorMessage string) {
	var updateEnvoyConfResponse = &UpdateEnvoyConfResponse{
		IsError:      true,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		ApiResponse:  EmptyApiResonse{},
	}
	json.NewEncoder(w).Encode(updateEnvoyConfResponse)
	w.WriteHeader(500)
}

func createSuccessResponse(w http.ResponseWriter) {
	var updateEnvoyConfResponse = &UpdateEnvoyConfResponse{
		IsError:      false,
		ErrorType:    "",
		ErrorMessage: "",
		ApiResponse:  EmptyApiResonse{},
	}
	json.NewEncoder(w).Encode(updateEnvoyConfResponse)
	w.WriteHeader(200)
}
