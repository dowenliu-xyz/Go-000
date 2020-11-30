package syncx

import "fmt"

func Go(x func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("%+v\n", err)
			}
		}()
		x()
	}()
}
