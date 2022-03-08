package game

import (
	"reflect"
	"sync"
	"testing"
)

func TestFindDiff(t *testing.T) {
	tree1 := generateTree(273, 5)
	tree2 := generateTree(273, 5, 213)

	c2v := make(chan ChallengerMessage, 100)
	v2p := make(chan ChallengerMessage, 100)
	p2v := make(chan ResponderMessage, 100)
	v2c := make(chan ResponderMessage, 100)

	c := &ChallengerSession{tree1, Hash{}, v2c, c2v}
	p := &ResponderSession{tree2, Hash{}, v2p, p2v}
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
		ToC:          v2c,
		FromC:        c2v,
		ToR:          v2p,
		FromR:        p2v,
		Dim:          5,
		MerkleHasher: NewSHA256Hasher(5),
	}
	msg := v.Run()

	diffData := msg.To
	diffData2 := tree2.nodes[tree2.leaves[213]].(inMemoryMerkleTreeLeaf).data
	if !reflect.DeepEqual(diffData, diffData2) {
		t.Error("responder sends incorrect leaf data")
	}
	diffPrev := msg.From
	diffPrev2 := tree2.nodes[tree2.leaves[212]].(inMemoryMerkleTreeLeaf).data
	if !reflect.DeepEqual(diffPrev, diffPrev2) {
		t.Error("responder sends incorrect leaf prev data")
	}

	wg.Wait()
}
