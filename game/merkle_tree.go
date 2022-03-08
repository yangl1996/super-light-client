package game

import (
	"crypto/sha256"
	"hash"
)

type Hash [32]byte

type MerkleTree interface {
	GetSubtreeSize() int
	GetRoots() []Hash
    GetChildren(node Hash) []Hash
    GetProof(node Hash) []Hash
	IsLeaf(node Hash) bool
	GetData(node Hash) interface{}
	GetPrevSibling(node Hash) Hash	// returns 0 if nonexistent
}

type MerkleHasher interface {
	ComputeParent(children []Hash) Hash
}

type SHA256Hasher struct {
	hasher hash.Hash
}

func NewSHA256Hasher() *SHA256Hasher {
	h := sha256.New()
	return &SHA256Hasher{h}
}

func (h *SHA256Hasher) ComputeParent(children []Hash) Hash {
	h.hasher.Reset()
	for _, c := range children {
		h.hasher.Write(c[:])
	}
	var res Hash
	h.hasher.Sum(res[:0])
	return res
}

type inMemoryMerkleTreeLeaf struct {
	data []byte
}

type inMemoryMerkleTreeInternal struct {
	children []Hash
	subtreeSize int
}

// a read-only merkle tree stored in the memory
type InMemoryMerkleTree struct {
	nodes map[Hash]interface{}
	leaves []Hash
	roots []Hash
	mh MerkleHasher
}

func NewInMemoryMerkleTree(data [][]byte, dim int) *InMemoryMerkleTree {
	mh := NewSHA256Hasher()
	m := &InMemoryMerkleTree{
		mh: mh,
		nodes: make(map[Hash]interface{}),
	}

	for len(data)> 0 {
		// compute the size of the next tree
		size := 1
		for size * dim <= len(data) {
			size = size * dim
		}
		var nextHashes []Hash
		for i := 0; i < size; i++ {
			l := inMemoryMerkleTreeLeaf{data[i]}
			h := sha256.Sum256(data[i])
			nextHashes = append(nextHashes, h)
			m.leaves = append(m.leaves, h)
			m.nodes[h] = l
		}
		for len(nextHashes) >= 1 {
			var hashes []Hash	// it is important that we allocate a new array because
								// internal nodes are referencing into nextHashes
			nb := len(nextHashes) / dim
			for i := 0; i < nb; i++ {
				n := inMemoryMerkleTreeInternal{
					children: nextHashes[nb*dim:nb*dim+dim],
					subtreeSize: size/nb,
				}
				h := m.mh.ComputeParent(nextHashes[nb*dim:nb*dim+dim])
				m.nodes[h] = n
				hashes = append(hashes, h)
			}
			nextHashes = hashes
		}
		// append the root
		m.roots = append(m.roots, nextHashes[0])

	}
	return m
}
