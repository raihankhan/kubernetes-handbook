package main

import (
	"fmt"
	"time"
)

func main() {
	lifeSpan := time.Minute * 10
	d := lifeSpan - time.Minute
	fmt.Println(d)
}
