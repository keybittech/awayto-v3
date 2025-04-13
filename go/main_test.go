package main

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
	"time"
)

func init() {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	go main()

	time.Sleep(2 * time.Second)

	statusCmd := mainApi.Handlers.Redis.RedisClient.FlushAll(context.Background())
	if statusCmd.Err() != nil {
		log.Fatal(statusCmd.Err())
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func Test_main(t *testing.T) {
	if !testing.Short() {
		t.Run("server runs for 5 seconds", func(t *testing.T) {
			time.Sleep(5 * time.Second)
		})
	}
}

// func BenchmarkBoolFormat(b *testing.B) {
// 	b.ReportAllocs()
// 	reset(b)
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
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		if false {
// 			_ = "t"
// 		}
// 	}
// }
