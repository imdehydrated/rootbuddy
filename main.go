package main

import (
	"log"
	"net/http"

	"github.com/imdehydrated/rootbuddy/server"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", server.NewServer()))
}
