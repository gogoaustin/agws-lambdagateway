package main

import (
	"log"
	"net/http"

	"git.gogoair.com/bagws/lambdagateway/app"
)

func main() {
	a := app.NewApp()

	log.Println("Listening on port :8000")
	log.Fatal(http.ListenAndServe(":8000", a))
}
