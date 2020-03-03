package main

import (
	"fmt"
)

type Func func()

func main() {
	Funcs := make([]Func, 0)
	for i := 0; i < 10; i++ {
		index := i
		Funcs = append(Funcs, func() {
			fmt.Println(index)
		})
	}
	for _, Func := range Funcs {
		Func()
	}
}
