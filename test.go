package main

import (
	"fmt"
	"sync/atomic"
)

var value uint64

func testAtomic() {
	foo := int(atomic.AddUint64(&value, uint64(1)) % uint64(5))
	fmt.Println("value is", value)
	fmt.Println("foo is", foo)
}

func main() {
	for i := 0; i < 6; i++ {
		testAtomic()
	}
}
