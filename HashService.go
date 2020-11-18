package main

import (
	"context"
	sha512 "crypto/sha512"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mapKeys int64              //used to increment the next key in the map
var totalTime int64            //used to measure the total time processing time all of the hashs took while the server was running
var hashMap map[int64]hashItem //stores the hashed items that represent timeHased and valueHashed
var myReport hashReport        //stores the preport that is updated and returned as a json document on the */stats request
var mutex = &sync.Mutex{}      //used to lock part of the hash function to only allow one thread to touch the map and modify the keys

//representation of the data being hashed using the hash imput phase, this does allow for duplicate value but different keys
type hashItem struct {
	TimeHashed time.Time
	HashValue  string
}

//used to represent the data that is returned in the json activity report
type hashReport struct {
	RequestCounts int64
	Average       int64
}

//main function
func main() {

	log.Println("HashServer starting")

	m := http.NewServeMux()

	//default if no server:port found
	server_port := "localhost:8080"

	if len(os.Args) == 2 {
		server_port = os.Args[1]
	}

	log.Println("HashServer using:", server_port)

	//use a property file configuration here to address the property but I would use
	//a non standard library. I chose not to implemented as the requirement restrict that usage
	s := http.Server{Addr: server_port, Handler: m}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mapKeys = 0
	hashMap = make(map[int64]hashItem)

	//represents the handlers to process the GET and POST request
	m.HandleFunc("/", handleHash)
	m.HandleFunc("/hash", handleHash)
	m.HandleFunc("/stats", reportStats)
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {

		// notify user that it received the shutdown message and is processing it accordingly
		w.Write([]byte("OK"))

		log.Println("HashServer shutting down")

		// Cancel the context on request
		cancel()
	})
	go func() {
		//this is a blocking call that listens and serves the request
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	//allows for a shutdown when context is cancelled
	select {
	case <-ctx.Done():

		s.Shutdown(ctx)
	}
}

//this function retuns the hashed value if the value exist and it has been in the map for at least 5 seconds
func getHashedValue(writer http.ResponseWriter, request *http.Request) {

	currentPath := strings.Split(request.URL.Path, "/")
	if len(currentPath) == 3 {

		hashKeyIn := currentPath[2]
		n, _ := strconv.Atoi(hashKeyIn)
		currentHashItem := hashMap[int64(n)]

		//hash value not available until after 5 seconds have passed since the item was hashed
		if 5 >= time.Now().Second()-currentHashItem.TimeHashed.Second() {

			message := []byte(currentHashItem.HashValue)
			_, err := writer.Write(message)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

//this handles hash request and either hashes the value our it retuns the user the value that was hash
func handleHash(writer http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case "GET":

		getHashedValue(writer, request)

	case "POST":

		currentPath := strings.Split(request.URL.Path, "/")

		if len(currentPath) == 2 {
			start := time.Now()

			sha_512 := sha512.New()

			//used to lock this part of the code for only a single thread
			mutex.Lock()
			pwd := request.FormValue("password")
			sha_512.Write([]byte(pwd))
			sEncoded := b64.StdEncoding.EncodeToString(sha_512.Sum(nil))
			mapKeys += 1
			newItem := hashItem{time.Now(), sEncoded}
			hashMap[mapKeys] = newItem

			message := []byte(strconv.FormatInt(mapKeys, 10))
			_, err := writer.Write(message)
			if err != nil {
				log.Fatal(err)
			}

			totalTime += time.Since(start).Microseconds() // converts time from nanoseconds to microseconds
			requestCnt := mapKeys

			myReport = hashReport{requestCnt, totalTime / requestCnt}

			//releases the lock and allows another thread to process
			defer mutex.Unlock()

		} else if len(currentPath) == 3 {

			getHashedValue(writer, request)

		} else {
			fmt.Fprintf(writer, "invalid URL")
		}
	default: // using this default to capture unused Methodes

		fmt.Fprintf(writer, "Only Only GET and POST methods are supported")
	}

}

//this function reports the values currently stored in the report that shows the counts and avgtime
func reportStats(writer http.ResponseWriter, request *http.Request) {

	err := json.NewEncoder(writer).Encode(myReport)
	if err != nil {
		log.Fatal(err)
	}
}
