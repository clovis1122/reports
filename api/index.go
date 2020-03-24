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
	fmt.Fprintf(w, report)
}

// func main() {
// 	http.HandleFunc("/api/report", Handler)
// 	http.ListenAndServe(":80", nil)
// }
