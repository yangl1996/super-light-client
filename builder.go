package main

import (
	"flag"
	"github.com/yangl1996/super-light-client/game"
	"log"
	"encoding/binary"
)

func buildTree(args []string) {
	cmd := flag.NewFlagSet("build", flag.ExitOnError)
	size := cmd.Int("size", 1000000, "number of elements to insert")
	path := cmd.String("file", "tree.pogreb", "file to store the dirty tree")
	dim := cmd.Int("dim", 50, "degree/dimension of the tree")
	diff := cmd.Int("diff", 0, "point of difference")
	cmd.Parse(args)

	testData := func(i int) []byte {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(i))
		if *diff != 0 && *diff == i {
			bs = append(bs, []byte("diff")...)
		}
		return bs
	}

	storage := game.NewPogrebMerkleTreeStorage(*path)
	game.NewKVMerkleTree(storage, testData, *size, *dim)
	log.Println("committing to the disk")
	storage.Commit()
	storage.Close()
}
