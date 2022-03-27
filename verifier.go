package main

import (
	"flag"
	"log"
	"github.com/yangl1996/super-light-client/game"
	"math"
	"net"
	"time"
	"sync"
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
	num := cmd.Int("N", 10, "number of back-to-back verifications per thread")
	burst := cmd.Int("p", 1, "number of threads to generate verifications")
	cmd.Parse(args)
	servers := cmd.Args()
	if len(servers) < 2 {
		log.Fatalln("supply at least 2 servers as command line arguments")
	}

	wg := &sync.WaitGroup{}
	l := &sync.Mutex{}
	tot := 0.0
	cnt := 0
	totSq := 0.0
	resCh := make(chan float64, 1000)
	l.Lock()
	go func() {
		defer l.Unlock()
		for s := range resCh {
			tot += s
			totSq += s * s
			cnt += 1
		}
	}()

	log.Printf("running verifications")
	initWg := &sync.WaitGroup{}
	for node := 0; node < *burst; node++ {
		wg.Add(1)
		initWg.Add(1)
		go func() {
			v := newVerifier(cmd.Args(), *deg)
			initWg.Done()
			initWg.Wait()
			for i := 0; i < *num; i++ {
				start := time.Now()
				_, winner := v.Run()
				dur := float64(time.Since(start).Milliseconds())
				resCh <- dur
				if *burst == 1 {
					log.Printf("server %v is winner\n", winner)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(resCh)
	l.Lock()
	avg := tot / float64(cnt)
	stddev := math.Sqrt(totSq / float64(cnt) - avg * avg)
	l.Unlock()

	log.Printf("finished %v runs, avg %.2f ms, stddev %.2f ms\n", cnt, avg, stddev)
}

