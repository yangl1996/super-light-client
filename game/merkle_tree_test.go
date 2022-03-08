package game

import (
	"encoding/binary"
	"testing"
)

func generateTree(sz, dim int, diff ...int) *InMemoryMerkleTree {
	nextDiffIdx := 0
	testData := [][]byte{}
	for i := 0; i < sz; i++ {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(i))
		if nextDiffIdx < len(diff) && diff[nextDiffIdx] == i {
			nextDiffIdx += 1
			bs = append(bs, []byte("diff")...)
		}
		testData = append(testData, bs)
	}
	return NewInMemoryMerkleTree(testData, dim)
}

func TestMerkleProof(t *testing.T) {
	m := generateTree(125, 5)
	p := m.GetProof(m.leaves[40])
	checker := NewSHA256Hasher(5)
	if !checker.CheckProof(m.nodes[m.leaves[40]].(inMemoryMerkleTreeLeaf).data, p, m.roots[0]) {
		t.Error("proof does not pass check")
	}
	m = generateTree(125, 5, 40)
	if checker.CheckProof(m.nodes[m.leaves[41]].(inMemoryMerkleTreeLeaf).data, p, m.roots[0]) {
		t.Error("incorrect proof passes check")
	}
}

func TestCreateMerkleTree(t *testing.T) {
	generateTree(94534, 7)
	generateTree(0, 2)
	generateTree(130, 2)

}
