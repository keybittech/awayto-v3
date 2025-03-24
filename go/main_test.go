package main

import (
	"flag"
	"log"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(5 * time.Second)
	println("did setup main")

}
