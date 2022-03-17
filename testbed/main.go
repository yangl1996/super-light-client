package main

import (
	"fmt"
	"os"
)

func helper() {
	fmt.Println("available commands: cluster, exp")
	os.Exit(1)
}

func main() {
	// dispatch subcommands
	if len(os.Args) <= 1 {
		helper()
	}
	switch os.Args[1] {
	case "cluster":
		dispatchCluster(os.Args[2:])
		return
	case "exp":
		dispatchBwTest(os.Args[2:])
		return
	default:
		helper()
	}
}
