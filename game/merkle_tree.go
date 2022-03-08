package game

import (
	"crypto/sha256"
	"hash"
)

type Hash [32]byte

type MerkleTree interface {
	GetSubtreeSize(node Hash) int
	GetRoots() []Hash
	GetChildren(node Hash) []Hash
	GetProof(node Hash) []Hash
	IsLeaf(node Hash) bool
	GetData(node Hash) []byte
	GetPrevSibling(node Hash) Hash // returns 0 if nonexistent
}

type MerkleHasher interface {
	HashData(data []byte) Hash
	ComputeParent(children []Hash) Hash
	CheckProof(leafData []byte, proof []Hash, roots ...Hash) bool
}

type SHA256Hasher struct {
	hasher hash.Hash
	dim    int
}

func NewSHA256Hasher(dim int) *SHA256Hasher {
	h := sha256.New()
	return &SHA256Hasher{h, dim}
}

func (h *SHA256Hasher) HashData(data []byte) Hash {
	r := Hash{}
	h.hasher.Reset()
	h.hasher.Write(data[:])
	h.hasher.Sum(r[:0])
	return r
}

func (h *SHA256Hasher) ComputeParent(children []Hash) Hash {
	if len(children) != h.dim {
		panic("incorrect dimension")
	}
	h.hasher.Reset()
	for _, c := range children {
		h.hasher.Write(c[:])
	}
	var res Hash
	h.hasher.Sum(res[:0])
	return res
}

func (m *SHA256Hasher) CheckProof(leafData []byte, proof []Hash, roots ...Hash) bool {
	leaf := m.HashData(leafData)
	for len(proof) > 0 {
		found := false
		// look for leaf in the next level
		for _, v := range proof[:m.dim] {
			if v == leaf {
				found = true
				break
			}
		}
		if !found {
			return false
		}
		leaf = m.ComputeParent(proof[:m.dim])
		proof = proof[m.dim:]
	}
	for _, r := range roots {
		if leaf == r {
			return true
		}
	}
	return false
}

type inMemoryMerkleTreeLeaf struct {
	data  []byte
	index int
}

type inMemoryMerkleTreeInternal struct {
	children    []Hash
	subtreeSize int
}

// a read-only merkle tree stored in the memory
type InMemoryMerkleTree struct {
	nodes  map[Hash]interface{}
	parent map[Hash]Hash
	leaves []Hash
	roots  []Hash
	mh     MerkleHasher
}

func (m *InMemoryMerkleTree) GetSubtreeSize(node Hash) int {
	switch n := m.nodes[node].(type) {
	case inMemoryMerkleTreeLeaf:
		return 1
	case inMemoryMerkleTreeInternal:
		return n.subtreeSize
	default:
		panic("unknown node type")
	}
}

func (m *InMemoryMerkleTree) GetRoots() []Hash {
	return m.roots
}

func (m *InMemoryMerkleTree) GetChildren(node Hash) []Hash {
	return m.nodes[node].(inMemoryMerkleTreeInternal).children
}

func (m *InMemoryMerkleTree) GetProof(node Hash) []Hash {
	_, yes := m.nodes[node].(inMemoryMerkleTreeLeaf)
	if !yes {
		panic("node is not a leaf")
	}
	proof := []Hash{}
	for {
		parent, there := m.parent[node]
		if !there {
			break
		}
		pn := m.nodes[parent].(inMemoryMerkleTreeInternal)
		proof = append(proof, pn.children...)
		node = parent
	}
	return proof
}

func (m *InMemoryMerkleTree) IsLeaf(node Hash) bool {
	switch m.nodes[node].(type) {
	case inMemoryMerkleTreeLeaf:
		return true
	case inMemoryMerkleTreeInternal:
		return false
	default:
		panic("unknown node type")
	}
}

func (m *InMemoryMerkleTree) GetData(node Hash) []byte {
	return m.nodes[node].(inMemoryMerkleTreeLeaf).data
}

func (m *InMemoryMerkleTree) GetPrevSibling(node Hash) Hash {
	n := m.nodes[node].(inMemoryMerkleTreeLeaf)
	if n.index > 0 {
		return m.leaves[n.index-1]
	} else {
		return Hash{}
	}
}

func NewInMemoryMerkleTree(data [][]byte, dim int) *InMemoryMerkleTree {
	mh := NewSHA256Hasher(dim)
	m := &InMemoryMerkleTree{
		mh:     mh,
		nodes:  make(map[Hash]interface{}),
		parent: make(map[Hash]Hash),
	}

	for len(data) > 0 {
		// compute the size of the next tree
		size := 1
		for size*dim <= len(data) {
			size = size * dim
		}
		var nextHashes []Hash
		for i := 0; i < size; i++ {
			l := inMemoryMerkleTreeLeaf{
				data:  data[i],
				index: len(m.leaves),
			}
			h := m.mh.HashData(data[i])
			nextHashes = append(nextHashes, h)
			m.leaves = append(m.leaves, h)
			m.nodes[h] = l
		}
		for len(nextHashes) > 1 {
			var hashes []Hash // it is important that we allocate a new array because
			// internal nodes are referencing into nextHashes
			nb := len(nextHashes) / dim
			for i := 0; i < nb; i++ {
				n := inMemoryMerkleTreeInternal{
					children:    nextHashes[i*dim : i*dim+dim],
					subtreeSize: size / nb,
				}
				h := m.mh.ComputeParent(nextHashes[i*dim : i*dim+dim])
				m.nodes[h] = n
				hashes = append(hashes, h)
				for j := 0; j < dim; j++ {
					m.parent[nextHashes[i*dim+j]] = h
				}
			}
			nextHashes = hashes
		}
		// append the root
		m.roots = append(m.roots, nextHashes[0])
		data = data[size:]
	}
	return m
}

var test MerkleTree = &InMemoryMerkleTree{}
