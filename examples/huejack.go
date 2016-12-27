package main

import (
	"fmt"
	"log"

	"github.com/pborges/huejack"
)

var lights = [...]string{
	"test",
	"kitchen",
	"living room",
	"dining room",
	"bedroom",
}

func main() {
	huejack.Handle(lights[:], func(key, val int) {
		fmt.Printf("setting light %v (%q) to %v\n", key, lights[key], val)
	})

	log.Fatal(huejack.ListenAndServe())
}
