package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewStore())
	log.Printf("notes-api %s listening on :8080", os.Getenv("APP_VERSION"))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
