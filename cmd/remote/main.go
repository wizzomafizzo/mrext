package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "App goes here.")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", home).Methods("GET")

	router.HandleFunc("/screenshots", allScreenshots).Methods("GET")
	router.HandleFunc("/screenshots", takeScreenshot).Methods("POST")
	router.HandleFunc("/screenshots/{core}/{image}", viewScreenshot).Methods("GET")
	router.HandleFunc("/screenshots/{core}/{image}", deleteScreenshot).Methods("DELETE")

	router.HandleFunc("/systems", allSystems).Methods("GET")
	router.HandleFunc("/systems/{id}", launchCore).Methods("POST")

	srv := &http.Server{
		// TODO: restrict this later
		Handler:      cors.AllowAll().Handler(router),
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
