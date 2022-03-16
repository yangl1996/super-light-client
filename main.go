package main

import (
	"fmt"
	"os"
	"math/rand"
	"time"
	"encoding/gob"
	"github.com/yangl1996/super-light-client/game"
)

func main() {
	gob.Register(game.OpenNext{})
	gob.Register(game.StartRoot{})
	gob.Register(game.NextChildren{})
	gob.Register(game.StateTransition{})
	gob.Register(game.MountainRange{})

	rand.Seed(time.Now().UnixNano())
	if len(os.Args) < 2 {
		fmt.Println("subcommands: verify, serve, build")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "verify":
		verify(os.Args[2:])
	case "serve":
		serve(os.Args[2:])
	case "build":
		buildTree(os.Args[2:])
	default:
		fmt.Println("unknown subcommand")
		os.Exit(1)
	}
}
