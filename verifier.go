package main

import (
	"flag"
	"log"
)

type peer struct {

}

func verify(args []string) {
	cmd := flag.NewFlagSet("verify", flag.ExitOnError)
	cmd.Parse(args)
	servers := cmd.Args()
	if len(servers) < 2 {
		log.Fatalln("supply at least 2 servers")
	}
}


