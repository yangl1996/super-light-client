package game

import (
	"reflect"
	"sync"
	"testing"
)

func TestFindDiff(t *testing.T) {
	tree1 := generateTree(273, 5)
	tree2 := generateTree(299, 5, 213)

	c2v := make(chan Message, 100)
	v2p := make(chan Message, 100)
	p2v := make(chan Message, 100)
	v2c := make(chan Message, 100)

	c := &Session{tree1, v2c, c2v, Hash{}}
	p := &Session{tree2, v2p, p2v, Hash{}}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.Run()
		wg.Done()
	}()
	go func() {
		p.Run()
		wg.Done()
	}()
	v := Verifier{
		To:           []chan<- Message{v2c, v2p},
		From:         []<-chan Message{c2v, p2v},
		Dim:          5,
		MerkleHasher: NewSHA256Hasher(5),
	}
	msg := v.Run()

	diffData := msg.To
	diffData2 := tree1.nodes[tree1.leaves[213]].(inMemoryMerkleTreeLeaf).data
	if !reflect.DeepEqual(diffData, diffData2) {
		t.Error("responder sends incorrect leaf data")
	}
	diffPrev := msg.From
	diffPrev2 := tree1.nodes[tree1.leaves[212]].(inMemoryMerkleTreeLeaf).data
	if !reflect.DeepEqual(diffPrev, diffPrev2) {
		t.Error("responder sends incorrect leaf prev data")
	}

	wg.Wait()
}
