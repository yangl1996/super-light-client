package game

import (
	"testing"
	"encoding/binary"
)

func generateTree(sz, dim int) *InMemoryMerkleTree {
	testData := [][]byte{}
	for i := 0; i < sz; i++ {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(i))
		testData = append(testData, bs)
	}
	return NewInMemoryMerkleTree(testData, dim)
}

func TestMerkleProof(t *testing.T) {
	m := generateTree(125, 5)
	p := m.GetProof(m.leaves[40])
	if !m.CheckProof(m.roots[0], m.leaves[40], p) {
		t.Error("proof does not pass check")
	}
	if m.CheckProof(m.roots[0], m.leaves[41], p) {
		t.Error("incorrect proof passes check")
	}

}
