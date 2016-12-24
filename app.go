package main

import (
	"log"
	"net/http"
	"strconv"
)

func main() {
	// An example of a HTTP endpoint at the server
	// which a client (browser) can make a dynamic HTTP call to.
	// This function will parse a number from the request, increment it, and send it back as the response.
	doServerStuff := func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters from request.
		queryParameters := r.URL.Query()

		// Get the number parameter we are looking for.
		numValues, ok := queryParameters["number"]
		if !ok {
			http.Error(w, "No number provided", http.StatusBadRequest)
			return
		}

		// Convert the number from string form into integer.
		num, err := strconv.Atoi(numValues[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Increment the number
		num++

		// Print new number to console and send it back to requester.
		log.Println(num)
		w.Write([]byte(strconv.Itoa(num)))
	}
	http.HandleFunc("/changenumber", doServerStuff) // Run this function when this URL path is hit.

	// Serve static files (files to be run in browser at client html, css, js) to browser when root URL is hit.
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Start HTTP server on localhost port 80 and exit when it fails.
	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":80", nil))
}
