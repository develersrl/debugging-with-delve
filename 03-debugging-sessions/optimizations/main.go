package main

func sum(a, b int) int {
	s := a + b

	return s
}

func number() int {
	return 10
}

func main() {
	a := 10
	b := 5

	c := sum(a, b)
	println(c)

	n := number()
	println(n)
}
