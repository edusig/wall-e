package main

import (
	"github.com/gorilla/mux"
)

func IndexHandler(res, req) {

}

func UploadHandler(res, req) {

}

func main() {

	router := mux.NewRouter()

	router.
		HandleFunc("/", IndexHandler).
		Methods("GET")

	router.
		HandleFunc("/index.html", IndexHandler).
		Methods("GET")

	router.
		HandleFunc("/upload", UploadHandler).
		Methods("POST")
}
