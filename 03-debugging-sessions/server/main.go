package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/develersrl/debugging-with-delve/03-debugging-sessions/server/isogram"
)

func main() {
	log.Println("listening on localhost:12345")
	log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	if isogram.IsIsogram(strings.TrimLeft(r.URL.Path, "/")) {
		fmt.Fprintln(w, "YES")
	} else {
		fmt.Fprintln(w, "NO")
	}
}
