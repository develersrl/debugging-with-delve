package main

import (
	"fmt"
	"sort"
)

type Counter interface {
	Next()
}

type Progression struct {
	Value int
	Step  int
}

func (p *Progression) Next() {
	p.Value += p.Step
}

func sum(a, b int) int {
	return a + b
}

func main() {
	a, b := 10, 5
	total := sum(a, b)
	fmt.Printf("%d + %d = %d\n", a, b, total)

	p := Progression{
		Value: 1,
		Step:  1,
	}

	for i := 0; i < 10; i++ {
		p.Next()
	}

	// Delve plays nice with interfaces...
	var c Counter = &p
	c.Next()

	// with slices...
	s := make([]int, 10)
	for i := 10; i > 0; i-- {
		s[10-i] = i
	}
	sort.Ints(s)

	// with maps...
	m := make(map[int]string)
	for i := 0; i < 10; i++ {
		m[i] = fmt.Sprintf("%d", i)
	}
	m[3] = "Debugging Workshop"
}
