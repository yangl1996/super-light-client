package game

import (
	"encoding/binary"
	"testing"
	//"os"
	//"path/filepath"
)

func generateTree(sz, dim int, diff ...int) *KVMerkleTree {
	nextDiffIdx := 0
	testData := func(i int) []byte {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(i))
		if nextDiffIdx < len(diff) && diff[nextDiffIdx] == i {
			nextDiffIdx += 1
			bs = append(bs, []byte("diff")...)
		}
		return bs
	}
	storage := NewInMemoryMerkleTreeStorage()
	//dir, err := os.MkdirTemp("", "merkle-test")
	//if err != nil {
	//	panic(err)
	//}
	//file := filepath.Join(dir, "db")
	//storage := NewPogrebMerkleTreeStorage(file)
	return NewKVMerkleTree(storage, testData, sz, dim)
}

func TestMerkleProof(t *testing.T) {
	m := generateTree(125, 5)
	p := m.GetProof(m.getLeafHashByIndex(40))
	checker := NewSHA256Hasher(5)

	n, _ := m.getLeaf(m.getLeafHashByIndex(40))
	if !checker.CheckProof(n.data, p, m.getRoot(0)) {
		t.Error("proof does not pass check")
	}
	m = generateTree(125, 5, 40)
	n, _ = m.getLeaf(m.getLeafHashByIndex(41))
	if checker.CheckProof(n.data, p, m.getRoot(0)) {
		t.Error("incorrect proof passes check")
	}
}

func TestCreateMerkleTree(t *testing.T) {
	generateTree(94534, 7)
	generateTree(0, 2)
	generateTree(130, 2)
}
