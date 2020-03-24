// package main

package handler

import (
	"fmt"
	"handler/reports"
	"net/http"
)

// Handler handles the request.
func Handler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	report, _ := reports.GetTogglReport(token)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, report)
}

// func main() {
// 	http.HandleFunc("/api", Handler)
// 	fs := http.FileServer(http.Dir("../"))
// 	http.Handle("/", fs)

// 	log.Println("Listening on :80...")
// 	err := http.ListenAndServe(":80", nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
