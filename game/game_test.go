package game

import (
	"crypto/sha256"
	"testing"
	"reflect"
	"sync"
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
	wg.Add(4)
	go func() {
		c.Run()
		wg.Done()
	}()
	go func() {
		p.Run()
		wg.Done()
	}()
	go func() {
		for m := range c2v {
			v2p <- m
		}
		wg.Done()
	}()
	go func() {
		stop := false
		for m := range p2v {
			switch msg := m.(type) {
			case NextChildren:
				if stop {
					t.Error("responder sends message after game finished")
				}
				v2c <- m
			case MountainRange:
				if stop {
					t.Error("responder sends message after game finished")
				}
				v2c <- m
			case StateTransition:
				stop = true
				diffData := msg.To.([]byte)
				diffData2 := tree2.nodes[tree2.leaves[213]].(inMemoryMerkleTreeLeaf).data
				if !reflect.DeepEqual(diffData, diffData2) {
					t.Error("responder sends incorrect leaf data")
				}
				diffPrev := msg.From.([]byte)
				diffPrev2 := tree2.nodes[tree2.leaves[212]].(inMemoryMerkleTreeLeaf).data
				if !reflect.DeepEqual(diffPrev, diffPrev2) {
					t.Error("responder sends incorrect leaf prev data")
				}
				checker := NewSHA256Hasher(5)
				if !checker.CheckProof(sha256.Sum256(diffPrev[:]), msg.FromProof, tree2.roots...) {
					t.Error("responder sends incorrect proof for leaf prev")
				}
			default:
				panic("unexpected data type")
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

