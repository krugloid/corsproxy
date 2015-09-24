package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type corsHandler struct{}

func (h *corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, r.URL.Path[1:]+"?"+r.URL.RawQuery, r.Body)
	for k, v := range r.Header {
		req.Header.Set(k, v[0])
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad request: ", err)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
	w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	port := ":8080"
	if s := os.Getenv("PORT"); s != "" {
		port = s
	}
	http.ListenAndServe(port, &corsHandler{})
}
