package main

import (
	"fmt"

	"github.com/Callmeden/go-cheatsheet/tasks/05-inmemorycache/cache"
)

func main() {
	c := cache.New()

	c.Set("userId", 42)

	if v, ok := c.Get("userId"); ok {
		fmt.Println(v) // 42
	}

	c.Delete("userId")

	if _, ok := c.Get("userId"); !ok {
		fmt.Println("not found")
	}
}
