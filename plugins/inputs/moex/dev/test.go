package main

import (
	"fmt"
)

var arr = []interface{}{
	"str", 1, nil,
}

func main() {
	fl, ok := arr[1].(int)
	fmt.Printf("%s %s", fl, ok)
}
