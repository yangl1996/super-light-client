package main

import (
	"flag"
	"log"
	"github.com/yangl1996/super-light-client/game"
	"net"
	"time"
)

func newVerifier(servers []string, deg int) *game.Verifier {
	var toProvers []chan<- game.Message
	var fromProvers []<-chan game.Message

	for _, addr := range servers {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		t := make(chan game.Message, 100)
		f := make(chan game.Message, 100)
		go readPeer(conn, f)
		go writePeer(conn, t)
		toProvers = append(toProvers, t)
		fromProvers = append(fromProvers, f)
	}

	v := game.Verifier {
		To: toProvers,
		From: fromProvers,
		Dim: deg,
		MerkleHasher: game.NewSHA256Hasher(deg),
	}
	return &v
}

func verify(args []string) {
	cmd := flag.NewFlagSet("verify", flag.ExitOnError)
	deg := cmd.Int("dim", 50, "dimension of the tree")
	num := cmd.Int("N", 100, "number of back-to-back verifications to perform")
	cmd.Parse(args)
	servers := cmd.Args()
	if len(servers) < 2 {
		log.Fatalln("supply at least 2 servers as command line arguments")
	}

	v := newVerifier(cmd.Args(), *deg)

	start := time.Now()
	for i := 0; i < *num; i++ {
		_, winner := v.Run()
		log.Printf("server %v is winner\n", winner)
	}
	dur := time.Since(start)
	log.Printf("time for %v verifications is %v us\n", *num, dur.Microseconds())
}


