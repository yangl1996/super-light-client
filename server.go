package main

import (
	"flag"
	"log"
)

func serve(args []string) {
	cmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := cmd.Int("port", 9000, "port to listen for incoming connections")
	cmd.Parse(args)
	log.Println(*port)
}
