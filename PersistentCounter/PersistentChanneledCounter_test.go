package PersistentCounter

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"time"
	"encoding/json"
	"math"
	"os"
	"log"
	"fmt"
	"errors"
	"strings"
	"strconv"
)

func WaitForNewSecond(timeStarted time.Time) (newTime time.Time) {
	for {
		secondsElapsed := math.Floor(time.Since(timeStarted).Seconds())
		if secondsElapsed > 0 {
			break
		}
	}
	return time.Now()
}

type RequestsCount struct {
	Requests uint64 `json:"Requests"`
}

func NewMockedPersistentServer(fileName string, persistSeconds uint64) (s* httptest.Server, err error) {
	pCounter, err := NewPersistentChanneledCounter(fileName, 50 * time.Millisecond, persistSeconds)
	if err != nil {
		return nil, err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r* http.Request){
		pCounter.IncreaseCounter()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"Requests": pCounter.GetTotals()})
		return
	}))

	return server, nil
}

func TestPersistentChanneledCounter_ResponseCorrectAfterOneSec(t *testing.T) {
	fileName := "test.txt"
	persistSeconds := 10
	const requestsCount = 10

	server, err := NewMockedPersistentServer(fileName, uint64(persistSeconds))

	defer server.Close()
	defer os.Remove(fileName)

	timeStarted := time.Now()

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	var data = RequestsCount{}

	// Send 10 requests
	for i := 0; i < requestsCount; i++ {
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err.Error())
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("Requests: %d", data)
	}

	timeStarted = WaitForNewSecond(timeStarted)
	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}


	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("Server response Requests: %d", data.Requests)
	if data.Requests != requestsCount {
		t.Fatalf("Incorrect number of responses: Expected: %d; Received %d",
			requestsCount, data.Requests)
	}
}

func TestPersistentChanneledCounter_ResponseCorrectAfterPeriod(t *testing.T) {
	fileName := "test.txt"
	persistSeconds := 5
	const requestsCount = 5

	server, err := NewMockedPersistentServer(fileName, uint64(persistSeconds))

	defer server.Close()
	defer os.Remove(fileName)

	timeStarted := time.Now()

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	var data = RequestsCount{}

	// Send <requestsCount> per sec for <persistSeconds> seconds.
	for seconds := 0; seconds < persistSeconds; seconds++ {

		for i := 0; i < requestsCount; i++ {
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err.Error())
			}

			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				t.Fatal(err.Error())
			}
		}
		timeStarted = WaitForNewSecond(timeStarted)
		t.Log("New second started")
	}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("Server response Requests: %d", data.Requests)
	if data.Requests != uint64(requestsCount * persistSeconds) {
		t.Fatalf("Incorrect number of responses: Expected: %d; Received %d",
			requestsCount * persistSeconds, data.Requests)
	}
}

func readDataFromFile(fileName string) (result []uint64, err error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	defer file.Close()

	if err != nil {
		return nil, fmt.Errorf("Unable to open file %s for reading and writing", fileName)
	}

	fileSize, err := file.Stat()
	if err != nil {
		return nil, errors.New("Unable to get file information")
	}

	buffer := make([]byte, fileSize.Size())
	file.Seek(0, 0)
	n, err := file.Read(buffer)
	log.Printf("%d bytes readed\r\n", n)
	if err != nil {
		return nil, errors.New("Unable to restore buffer from file")
	}

	stringData := strings.Split(string(buffer), ",")
	if len(stringData) == 0 {
		return nil, errors.New("Corrupted file")
	}

	var res = make([]uint64, len(stringData))
	for _, s := range stringData {
		d, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, errors.New("Unable to parse file")
		}
		res = append(res, d)
	}

	return res, nil
}

func saveDataToFile(fileName string, data []uint64) (err error) {
	file, err := os.OpenFile(fileName, os.O_CREATE | os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("Unable to open file %s for reading and writing", fileName)
	}
	file.Truncate(0)
	file.Seek(0, 0)
	stringData := make([]string, len(data))
	for i := 0; i < len(data); i++ {
		stringData[i] = strconv.Itoa(int(data[i]))
	}
	result := strings.Join(stringData, ",")
	file.Write([]byte(result))
	return nil
}

func TestPersistentChanneledCounter_PersistFileAfterOneRequest(t *testing.T) {
	fileName := "test.txt"
	persistSeconds := 5

	server, err := NewMockedPersistentServer(fileName, uint64(persistSeconds))

	defer server.Close()
	defer os.Remove(fileName)

	timeStarted := time.Now()

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	var data = RequestsCount{}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")
	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	fileData, err := readDataFromFile(fileName)
	if err != nil {
		t.Fatal(err.Error())
	}

	var sum uint64 = 0
	t.Logf("Slice size: %d", len(fileData))
	for _, f := range fileData {
		sum += f
		t.Logf("Value: %d", f)
	}


	if sum != 1 {
		log.Fatal("Incorrect counter value")
	}
}


func TestPersistentChanneledCounter_PersistFileAfterPeriod(t *testing.T) {
	fileName := "test.txt"
	persistSeconds := 5
	const requestsCount = 5

	server, err := NewMockedPersistentServer(fileName, uint64(persistSeconds))

	defer server.Close()
	defer os.Remove(fileName)

	timeStarted := time.Now()

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	timeStarted = WaitForNewSecond(timeStarted)
	t.Log("New second started")

	// Send <requestsCount> per sec for <persistSeconds> seconds.
	for seconds := 0; seconds < persistSeconds; seconds++ {

		for i := 0; i < requestsCount; i++ {
			_, err := client.Do(req)
			if err != nil {
				t.Fatal(err.Error())
			}
		}
		timeStarted = WaitForNewSecond(timeStarted)
		t.Log("New second started")
	}

	//timeStarted = WaitForNewSecond(timeStarted)
	//t.Log("New second started")
	//timeStarted = WaitForNewSecond(timeStarted)
	//t.Log("New second started")

	fileData, err := readDataFromFile(fileName)
	if err != nil {
		t.Fatal(err.Error())
	}

	var sum uint64 = 0
	t.Logf("Slice size: %d", len(fileData))
	for _, f := range fileData {
		sum += f
		t.Logf("Value: %d", f)
	}
	expected := uint64(persistSeconds * requestsCount)
	t.Logf("Total requests from file: %d", sum)
	t.Logf("Expected requests: %d", expected)

	if sum != expected {
		log.Fatal("Incorrect counter value")
	}
}

func TestPersistentChanneledCounter_PersistCounterRestoreFromFile(t *testing.T) {
	fileName := "test.txt"
	persistSeconds := 5

	var sampleData = []uint64{1, 2, 3, 4, 5}

	saveDataToFile(fileName, sampleData)

	server, err := NewMockedPersistentServer(fileName, uint64(persistSeconds))

	defer server.Close()
	defer os.Remove(fileName)

	req, err := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}

	var data = RequestsCount{}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	var sum uint64 = 0
	for _, d := range sampleData {
		sum += d
	}

	t.Logf("Requests: %d", data.Requests)
	t.Logf("Summ of samples: %d", sum)
	if data.Requests != sum {
		log.Fatal("Incorrect counter value")
	}
}