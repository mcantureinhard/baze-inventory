package main

import (
	"inventory/service/src/datastore"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

func main() {
	// Get the "PORT" env variable
	//port := os.Getenv("PORT")
	port := "8081"

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := datastore.NewRouter() // create routes

	// These two lines are important if you're designing a front-end to utilise this API methods
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"})

	// Launch server with CORS validations
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(allowedOrigins, allowedMethods)(router)))
}
