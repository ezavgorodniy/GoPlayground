package main

import (
	"net/http"
	"net/url"
)
import "io"

func handler(w http.ResponseWriter, r *http.Request) {
	escaped := url.QueryEscape(r.URL.Query().Get("param1"))
	io.WriteString(w, escaped)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
