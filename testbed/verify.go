package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func dispatchVerify(args []string) {
	command := flag.NewFlagSet("verify", flag.ExitOnError)
	serverListFilePath := command.String("l", "servers.json", "path to the server list file")

	command.Parse(args[0:])

	if *serverListFilePath == "" {
		fmt.Println("missing server list")
		os.Exit(1)
	}

	// parse the server list
	servers := ReadServerInfo(*serverListFilePath)
	addrs := []string{}

	for _, s := range servers {
		addrs = append(addrs, fmt.Sprintf("%s:9000", s.PublicIP))
	}

	fmt.Println(strings.Join(addrs, " "))

}
