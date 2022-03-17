package main

import (
	"flag"
	"log"
	"github.com/yangl1996/super-light-client/game"
	"math"
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
	num := cmd.Int("N", 10, "number of back-to-back verifications to perform")
	cmd.Parse(args)
	servers := cmd.Args()
	if len(servers) < 2 {
		log.Fatalln("supply at least 2 servers as command line arguments")
	}

	v := newVerifier(cmd.Args(), *deg)

	durs := []float64{}
	log.Printf("running verifications")
	for i := 0; i < *num; i++ {
		start := time.Now()
		_, winner := v.Run()
		dur := float64(time.Since(start).Milliseconds())
		durs = append(durs, dur)
		log.Printf("server %v is winner\n", winner)
	}
	avg := 0.0
	sqavg := 0.0
	for _, dur := range durs {
		avg += dur
		sqavg += dur * dur
	}
	avg /= float64(len(durs))
	sqavg /= float64(len(durs))

	log.Printf("finished %v runs, avg %.2f ms, stddev %.2f ms\n", len(durs), avg, math.Sqrt(sqavg-avg*avg))
}

