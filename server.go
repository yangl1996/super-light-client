package main

import (
	"flag"
	"log"
	"net"
	"encoding/gob"
	"github.com/yangl1996/super-light-client/game"
)

func serve(args []string) {
	cmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := cmd.String("addr", ":9000", "addr to listen for incoming connections")
	dbPath := cmd.String("db", "tree.pogreb", "path to the database file")
	cmd.Parse(args)

	db := game.NewPogrebMerkleTreeStorage(*dbPath)
	tree := game.OpenKVMerkleTree(db)

	l, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn, tree)
	}
}

func handleConn(conn net.Conn, tree game.MerkleTree) error {
	toPeer := make(chan game.Message, 100)
	fromPeer := make(chan game.Message, 100)
	go writePeer(conn, toPeer)
	go readPeer(conn, fromPeer)
	s := &game.Session{
		Tree: tree,
		I: fromPeer,
		O: toPeer,
	}
	s.Run()
	return nil
}

func readPeer(conn net.Conn, ch chan<- game.Message) error {
	defer close(ch)
	dec := gob.NewDecoder(conn)
	for {
		var d game.Message
		if err := dec.Decode(&d); err != nil {
			return err
		}
		ch <- d
	}
	return nil
}

func writePeer(conn net.Conn, ch <-chan game.Message) error {
	enc := gob.NewEncoder(conn)
	var err error
	for m := range ch {
		if err != nil {
			continue
		}
		err = enc.Encode(&m)
	}
	return err
}
