package game

import (
	"reflect"
	"sync"
	"testing"
)

func TestFindDiff(t *testing.T) {
	tree1 := generateTree(273, 5)
	tree2 := generateTree(299, 5, 271)

	p1v := make(chan Message, 100)
	p2v := make(chan Message, 100)
	vp1 := make(chan Message, 100)
	vp2 := make(chan Message, 100)

	p1 := &Session{tree1, vp1, p1v, Hash{}}
	p2 := &Session{tree2, vp2, p2v, Hash{}}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		p1.Run()
		wg.Done()
	}()
	go func() {
		p2.Run()
		wg.Done()
	}()
	v := Verifier{
		To:           []chan<- Message{vp1, vp2},
		From:         []<-chan Message{p1v, p2v},
		Dim:          5,
		MerkleHasher: NewSHA256Hasher(5),
	}
	mr := v.Run()
	if !reflect.DeepEqual(mr, p1.mountainRange()) {
		// player 1 should win, because we do not check state transition for now, and it plays by the rule all the time
		t.Error("incorrect winner")
	}

	wg.Wait()
}
