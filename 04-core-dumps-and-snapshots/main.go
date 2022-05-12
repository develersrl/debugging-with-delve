package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	log.Println("listening on localhost:12345")
	log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))

}

func handler(w http.ResponseWriter, r *http.Request) {
	if strings.TrimLeft(r.URL.Path, "/") == "crash" {
		go func() {
			panic(":(")
		}()
	}
	time.Sleep(500 * time.Millisecond)
	fmt.Fprintf(w, "hello from %d\n", os.Getpid())
}
