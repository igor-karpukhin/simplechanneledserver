package PersistentCounter

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"time"
	"encoding/json"
	"log"
)

func waitForNewSecond(timeStarted time.Time) (newTime time.Time) {
	for {
		secondsElapsed := time.Since(timeStarted).Seconds()
		if secondsElapsed > 0 {
			break
		}
	}
	return time.Now()
}

type RequestsCount struct {
	Requests uint64 `json:"Requests"`
}

func TestPersistentChanneledCounter_ResponseCorrect(t *testing.T) {
	persistentCounter, err := NewPersistentChanneledCounter("test.txt", 50 * time.Millisecond,
		10)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r* http.Request){
		persistentCounter.IncreaseCounter()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"Requests": persistentCounter.GetTotals()})
		return
	}))

	timeStarted := time.Now()

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	timeStarted = waitForNewSecond(timeStarted)

	const requestsCount = 10
	// Send 10 requests
	for i := 0; i < requestsCount; i++ {
		if _, err := client.Do(req); nil != err {
			t.Fatal(err.Error())
		}
	}

	timeStarted = waitForNewSecond(timeStarted)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}

	var data = RequestsCount{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatal(err.Error())
	}

	if data.Requests != requestsCount {
		log.Fatalf("Incorrect number of responses: Expected: %d; Received %d",
			requestsCount, data.Requests)
	}
}