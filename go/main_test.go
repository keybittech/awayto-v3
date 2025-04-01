package main

import (
	"flag"
	"log"
	"time"
)

func init() {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(1 * time.Second)
	println("did setup main")

	setupSockServer()
}
