package main

import (
	"flag"
	"log"
	"testing"
	"time"
)

func init() {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(5 * time.Second)
	println("did setup main")

	setupSockServer()
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

// func BenchmarkBoolFormat(b *testing.B) {
// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for c := 0; c < b.N; c++ {
//
// 		// For true
// 		var v bool = true
// 		var expected rune = 't'
// 		var actual rune
//
// 		// Zero-allocation way to get first char of bool
// 		if v {
// 			actual = 't'
// 		} else {
// 			actual = 'f'
// 		}
//
// 		if actual != expected {
// 			b.Fail()
// 		}
// 	}
// }
//
// func BenchmarkBoolAllocate(b *testing.B) {
// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for c := 0; c < b.N; c++ {
// 		if false {
// 			_ = "t"
// 		}
// 	}
// }
