package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)


type postArgs struct {
	Min float64 `json:"min" validate:"required,gte=0,lte=10"`
	Max float64 `json:"max" validate:"gte=100,lte=110"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var args postArgs
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unable to decode request body: %v", err)
		return
	}

	validate := validator.New()
	err = validate.Struct(args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Input is invalid: %v", err)
		return
	}

	io.WriteString(w,"OK")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
