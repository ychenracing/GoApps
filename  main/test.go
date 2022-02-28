package main

import (
	"fmt"
	"time"
)

func main() {

	s := time.Now()

	time.Sleep(time.Second * 2)
	n := time.Since(s)
	fmt.Println(n.Seconds())

}
