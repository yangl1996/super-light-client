package main

import (
	"flag"
	"log"
)

func main() {
	listenPort := flag.Int("port", 8000, "port to listen to")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())	// this is for hashes so seeds do not matter

	m.Mine()
}

)
