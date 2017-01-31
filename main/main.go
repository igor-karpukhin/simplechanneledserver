package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"runtime"
	"time"

	"github.com/igor-karpukhin/SimpleChanneledServer"
)

var (
	addr         = flag.String("addr", ":9090", "Address to bind")
	fileName     = flag.String("file", "storage.txt", "File to save/read data to/form")
	storeSeconds = flag.String("seconds", "60", "Number of seconds to store")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	seconds, err := strconv.ParseUint(*storeSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	persistentCounter, err := SimpleChanneledServer.NewPersistentChanneledCounter(
		*fileName, time.Duration(10)*time.Millisecond, seconds)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//log.Println("Incoming request")
		persistentCounter.IncreaseCounter()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"Requests": persistentCounter.GetTotals()})
		return
	})
	log.Println("Server started at:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
