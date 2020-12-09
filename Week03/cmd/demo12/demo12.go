package main

import "fmt"

type IceCreamMaker interface {
	// Hello greets a customer
	Hello()
}

// [1]
//type Ben struct {
//	name string
//}

// [2]
//type Ben struct {
//	id   int
//	name string
//}

// [3]
type Ben struct {
	name   *[5]byte
	field2 int
}

func (b *Ben) Hello() {
	fmt.Printf("Ben says, \"Hello my name is %v\"\n", b.name)
}

type Jerry struct {
	name string
}

func (j *Jerry) Hello() {
	fmt.Printf("Jerry says, \"Hello my name is %s\"\n", j.name)
}

// [1] 和 [3] 的内存布局相同，[2] 与 [1] 和 [3] 都不同。
// [2] 会以 panic 结束。

func main() {
	// [1]
	//var ben = &Ben{
	//	name: "Ben",
	//}
	// [2]
	//var ben = &Ben{
	//	id:   10,
	//	name: "Ben",
	//}
	// [3]
	var ben = &Ben{
		name:   new([5]byte),
		field2: 1,
	}
	var jerry = &Jerry{"Jerry"}
	var maker IceCreamMaker = ben

	var loop0, loop1 func()

	loop0 = func() {
		maker = ben
		go loop1()
	}

	loop1 = func() {
		maker = jerry
		go loop0()
	}

	go loop0()

	for {
		maker.Hello()
	}
}
