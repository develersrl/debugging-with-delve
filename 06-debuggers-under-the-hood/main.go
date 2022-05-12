package main

import "fmt"

//go:noinline
func BreakMe() {
	fmt.Println("Hello, Advanced Go Course")
}

func main() {
	BreakMe()
}
