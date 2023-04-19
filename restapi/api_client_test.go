package restapi

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"
)

type FollowResponse struct {
	RedirectURI string
}

type AsyncResponse struct {
	Status string
}

var apiClientServer *http.Server

func TestAPIClient(t *testing.T) {
	debug := true

	if debug {
		log.Println("client_test.go: Starting HTTP server")
	}
	setupAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:                 "http://127.0.0.1:8083/",
		insecure:            false,
		username:            "",
		password:            "",
		headers:             make(map[string]string),
		timeout:             2,
		idAttribute:         "id",
		copyKeys:            make([]string, 0),
		writeReturnsObject:  false,
		createReturnsObject: false,
		rateLimit:           1,
		debug:               debug,
	}
	client, _ := NewAPIClient(opt)

	var res string
	var err error

	if debug {
		log.Printf("api_client_test.go: Testing standard OK request\n")
	}
	res, err = client.sendRequest("GET", "/ok", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	if debug {
		log.Printf("api_client_test.go: Testing redirect request\n")
	}
	res, err = client.sendRequest("GET", "/redirect", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify if async calls work */
	if debug {
		log.Println("api_client_test.go: Testing asynchronous API call")
	}

	clientWithAsyncOptions := client
	clientWithAsyncOptions.AsyncSettings = &AsyncSettings{
		SearchKey:   "Status",
		SearchValue: "Done",
	}

	res, err = clientWithAsyncOptions.sendRequest("GET", "/async", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "{\"Status\":\"Done\"}\n" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify if redirect works */
	clientWithAsyncOptions.AsyncSettings = &AsyncSettings{RedirectUriKey: "RedirectURI"}

	if debug {
		log.Printf("api_client_test.go: Testing redirect request\n")
	}
	res, err = client.sendRequest("GET", "/custom-redirect", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify if async redirect works */
	clientWithAsyncOptions.AsyncSettings = &AsyncSettings{
		RedirectUriKey: "RedirectURI",
		SearchKey:      "Status",
		SearchValue:    "Done",
	}

	if debug {
		log.Printf("api_client_test.go: Testing async redirect request\n")
	}
	res, err = client.sendRequest("GET", "/custom-async-redirect", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "{\"Status\":\"Done\"}\n" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify timeout works */
	if debug {
		log.Printf("api_client_test.go: Testing timeout aborts requests\n")
	}
	_, err = client.sendRequest("GET", "/slow", "")
	if err == nil {
		t.Fatalf("client_test.go: Timeout did not trigger on slow request")
	}

	if debug {
		log.Printf("api_client_test.go: Testing rate limited OK request\n")
	}
	startTime := time.Now().Unix()

	for i := 0; i < 4; i++ {
		client.sendRequest("GET", "/ok", "")
	}

	duration := time.Now().Unix() - startTime
	if duration < 3 {
		t.Fatalf("client_test.go: requests not delayed\n")
	}

	if debug {
		log.Println("client_test.go: Stopping HTTP server")
	}
	shutdownAPIClientServer()
	if debug {
		log.Println("client_test.go: Done")
	}
}

func setupAPIClientServer() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It works!"))
	})
	serverMux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(9999 * time.Second)
		w.Write([]byte("This will never return!!!!!"))
	})
	serverMux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusPermanentRedirect)
	})
	serverMux.HandleFunc("/custom-redirect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		responseData := FollowResponse{RedirectURI: "http://127.0.0.1:8083/ok"}
		json.NewEncoder(w).Encode(responseData)
	})
	serverMux.HandleFunc("/custom-async-redirect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		responseData := FollowResponse{RedirectURI: "http://127.0.0.1:8083/async"}
		json.NewEncoder(w).Encode(responseData)
	})
	var counter = 0
	serverMux.HandleFunc("/async", func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.Header().Set("Content-Type", "application/json")
		var status string

		if counter <= 2 {
			status = "Pending"
		} else {
			status = "Done"
			counter = 0
		}

		responseData := AsyncResponse{Status: status}
		json.NewEncoder(w).Encode(responseData)
	})

	apiClientServer = &http.Server{
		Addr:    "127.0.0.1:8083",
		Handler: serverMux,
	}
	go apiClientServer.ListenAndServe()
	/* let the server start */
	time.Sleep(1 * time.Second)
}

func shutdownAPIClientServer() {
	apiClientServer.Close()
}
