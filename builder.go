package main

import (
	"flag"
	"github.com/yangl1996/super-light-client/game"
//	"log"
)

func buildTree(args []string) {
	cmd := flag.NewFlagSet("build", flag.ExitOnError)
	size := flag.Int("size", 1000000, "number of elements to insert")
	path := flag.String("file", "tree.pogreb", "file to store the dirty tree")
	dim := flag.Int("dim", 50, "degree/dimension of the tree")
	cmd.Parse(args)

	storage := game.NewPogrebMerkleTreeStorage(*path)
	game.NewRandomKVMerkleTree(storage, *size, *dim)
	storage.Commit()
	storage.Close()
}
