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
	GetData(node Hash) interface{}
	GetPrevSibling(node Hash) Hash	// returns 0 if nonexistent
	CheckProof(root Hash, leaf Hash, proof []Hash) bool
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
	index int
}

type inMemoryMerkleTreeInternal struct {
	children []Hash
	subtreeSize int
}

// a read-only merkle tree stored in the memory
type InMemoryMerkleTree struct {
	nodes map[Hash]interface{}
	parent map[Hash]Hash
	leaves []Hash
	roots []Hash
	mh MerkleHasher
	dim int
}

func (m *InMemoryMerkleTree) CheckProof(root Hash, leaf Hash, proof []Hash) bool {
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
		leaf = m.mh.ComputeParent(proof[:m.dim])
		proof = proof[m.dim:]
	}
	if leaf == root {
		return true
	} else {
		return false
	}
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

func (m *InMemoryMerkleTree) GetData(node Hash) interface{} {
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
	mh := NewSHA256Hasher()
	m := &InMemoryMerkleTree{
		mh: mh,
		nodes: make(map[Hash]interface{}),
		parent: make(map[Hash]Hash),
		dim: dim,
	}

	for len(data)> 0 {
		// compute the size of the next tree
		size := 1
		for size * dim <= len(data) {
			size = size * dim
		}
		var nextHashes []Hash
		for i := 0; i < size; i++ {
			l := inMemoryMerkleTreeLeaf{
				data: data[i],
				index: len(m.leaves),
			}
			h := sha256.Sum256(data[i])
			nextHashes = append(nextHashes, h)
			m.leaves = append(m.leaves, h)
			m.nodes[h] = l
		}
		for len(nextHashes) > 1 {
			var hashes []Hash	// it is important that we allocate a new array because
								// internal nodes are referencing into nextHashes
			nb := len(nextHashes) / dim
			for i := 0; i < nb; i++ {
				n := inMemoryMerkleTreeInternal{
					children: nextHashes[i*dim:i*dim+dim],
					subtreeSize: size/nb,
				}
				h := m.mh.ComputeParent(nextHashes[i*dim:i*dim+dim])
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
