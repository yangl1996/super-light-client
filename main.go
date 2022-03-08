package main

import (
	"fmt"
	"os"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) < 2 {
		fmt.Println("subcommands: verify, serve")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "verify":
		verify(os.Args[2:])
	case "serve":
		serve(os.Args[2:])
	}
}
