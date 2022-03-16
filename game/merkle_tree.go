package game

import (
	"crypto/sha256"
	"github.com/akrylysov/pogreb"
	"hash"
	"encoding/binary"
	"encoding"
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

type kvMerkleTreeLeaf struct {
	data  []byte
	index int
}

type kvMerkleTreeInternal struct {
	children    []Hash
	subtreeSize int
}

type kvMerkleTreeStorage interface {
	getLeaf(h Hash) (kvMerkleTreeLeaf, bool)
	getLeafHashByIndex(idx int) Hash
	getInternal(h Hash) (kvMerkleTreeInternal, bool)
	getParent(h Hash) (Hash, bool)
	getRoot(idx int) Hash
	getNumRoots() int
	appendLeaf(h Hash, l kvMerkleTreeLeaf)
	storeInternal(h Hash, n kvMerkleTreeInternal)
	storeParent(child Hash, parent Hash)
	appendRoot(h Hash)
}

type PogrebMerkleTreeStorage struct {
	db *pogreb.DB
}

func (s *PogrebMerkleTreeStorage) readObjectByHash(prefix [8]byte, h Hash, ret encoding.BinaryUnmarshaler) bool {
	key := [40]byte{}
	copy(key[0:8], prefix[:])
	copy(key[8:40], h[:])
	val, err := s.db.Get(key[:])
	if err != nil {
		return false
	}
	err = ret.UnmarshalBinary(val)
	if err != nil {
		panic(err)
	}
	return true
}

func (s *PogrebMerkleTreeStorage) writeObjectByHash(prefix [8]byte, h Hash, v encoding.BinaryMarshaler) {
	key := [40]byte{}
	copy(key[0:8], prefix[:])
	copy(key[8:40], h[:])
	buf, err := v.MarshalBinary()
	if err != nil {
		panic(err)
	}
	err = s.db.Put(key[:], buf[:])
	if err != nil {
		panic(err)
	}
	return
}

func (s *PogrebMerkleTreeStorage) readHashByIndex(prefix [8]byte, idx uint64) (Hash, bool) {
	key := [16]byte{}
	copy(key[0:8], prefix[:])
	binary.LittleEndian.PutUint64(key[8:], idx)
	val, err := s.db.Get(key[:])
	if err != nil {
		return Hash{}, false
	}
	var res Hash
	copy(res[:], val[0:32])
	return res, true
}

func (s *PogrebMerkleTreeStorage) writeHashByIndex(prefix [8]byte, idx uint64, v Hash) {
	key := [16]byte{}
	copy(key[0:8], prefix[:])
	binary.LittleEndian.PutUint64(key[8:], idx)
	err := s.db.Put(key[:], v[:])
	if err != nil {
		panic(err)
	}
	return
}

func (s *PogrebMerkleTreeStorage) readHashByHash(prefix [8]byte, h Hash) (Hash, bool) {
	key := [40]byte{}
	copy(key[0:8], prefix[:])
	copy(key[8:40], h[:])
	val, err := s.db.Get(key[:])
	if err != nil {
		return Hash{}, false
	}
	var res Hash
	copy(res[:], val[0:32])
	return res, true
}

func (s *PogrebMerkleTreeStorage) writeHashByHash(prefix [8]byte, k Hash, v Hash) {
	key := [40]byte{}
	copy(key[0:8], prefix[:])
	copy(key[8:40], k[:])
	err := s.db.Put(key[:], v[:])
	if err != nil {
		panic(err)
	}
	return
}

func (s *PogrebMerkleTreeStorage) readUint64(key [8]byte) uint64 {
	val, err := s.db.Get(key[:])
	if err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(val[:])
}

func (s *PogrebMerkleTreeStorage) writeUint64(key [8]byte, d uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], d)
	err := s.db.Put(key[:], buf[:])
	if err != nil {
		panic(err)
	}
	return
}

var leafNodePrefix = [8]byte{1}
var internalNodePrefix = [8]byte{2}
var leafHashPrefix = [8]byte{3}
var parentHashPrefix = [8]byte{4}
var rootHashPrefix = [8]byte{5}
var numberOfRootPrefix = [8]byte{6}
var numberOfLeafPrefix = [8]byte{7}

func (s *PogrebMerkleTreeStorage) getLeaf(h Hash) (kvMerkleTreeLeaf, bool) {
	var res kvMerkleTreeLeaf
	ok := s.readObjectByHash(leafNodePrefix, h, &res)
	return res, ok
}

func (s *PogrebMerkleTreeStorage) getLeafHashByIndex(idx int) Hash {
	hash, ok := s.readHashByIndex(leafHashPrefix, uint64(idx))
	if !ok {
		panic("index does not exist")
	}
	return hash
}

func (s *PogrebMerkleTreeStorage) getInternal(h Hash) (kvMerkleTreeInternal, bool) {
	var res kvMerkleTreeInternal
	ok := s.readObjectByHash(internalNodePrefix, h, &res)
	return res, ok
}

func (s *PogrebMerkleTreeStorage) getParent(h Hash) (Hash, bool) {
	return s.readHashByHash(parentHashPrefix, h)
}

func (s *PogrebMerkleTreeStorage) getRoot(idx int) Hash {
	hash, ok := s.readHashByIndex(parentHashPrefix, uint64(idx))
	if !ok {
		panic("index does not exist")
	}
	return hash
}

func (s *PogrebMerkleTreeStorage) getNumRoots() int {
	return s.readUint64(numberOfRootPrefix)
}

func (s *PogrebMerkleTreeStorage) appendLeaf(h Hash, l kvMerkleTreeLeaf) {
	idx := s.readUint64(numberOfLeafPrefix)
	s.writeObjectByHash(leafNodePrefix, h, l)
	s.writeHashByIndex(leafHashPrefix, idx, h)
	s.writeUint64(numberOfLeafPrefix, idx+1)
	return
}
func (s *PogrebMerkleTreeStorage) storeInternal(h Hash, n kvMerkleTreeInternal) {
	s.writeObjectByHash(internalNodePrefix, h, n)
	return
}

func (s *PogrebMerkleTreeStorage) storeParent(child Hash, parent Hash) {
	s.writeHashByHash(parentHashPrefix, child, parent)
	return
}

func (s *PogrebMerkleTreeStorage) appendRoot(h Hash) {
	idx := s.readUint64(numberOfRootPrefix)
	s.writeHashByIndex(rootHashPrefix, idx, h)
	s.writeUint64(numberOfRootPrefix, idx+1)
	return
}

// a read-only merkle tree stored in the memory
type InMemoryMerkleTreeStorage struct {
	nodes  map[Hash]interface{}
	parent map[Hash]Hash
	leaves []Hash
	roots  []Hash
}

func NewInMemoryMerkleTreeStorage() *InMemoryMerkleTreeStorage {
	return &InMemoryMerkleTreeStorage {
		nodes: make(map[Hash]interface{}),
		parent: make(map[Hash]Hash),
	}
}

func (s *InMemoryMerkleTreeStorage) getLeaf(h Hash) (kvMerkleTreeLeaf, bool) {
	switch n := s.nodes[h].(type) {
	case kvMerkleTreeLeaf:
		return n, true
	default:
		return kvMerkleTreeLeaf{}, false
	}
}

func (s *InMemoryMerkleTreeStorage) getLeafHashByIndex(idx int) Hash {
	return s.leaves[idx]
}

func (s *InMemoryMerkleTreeStorage) getInternal(h Hash) (kvMerkleTreeInternal, bool) {
	switch n := s.nodes[h].(type) {
	case kvMerkleTreeInternal:
		return n, true
	default:
		return kvMerkleTreeInternal{}, false
	}
}

func (s *InMemoryMerkleTreeStorage) getParent(h Hash) (Hash, bool) {
	data, ok := s.parent[h]
	return data, ok
}

func (s *InMemoryMerkleTreeStorage) getRoot(idx int) Hash {
	return s.roots[idx]
}

func (s *InMemoryMerkleTreeStorage) getNumRoots() int {
	return len(s.roots)
}

func (s *InMemoryMerkleTreeStorage) appendLeaf(h Hash, l kvMerkleTreeLeaf) {
	s.nodes[h] = l
	s.leaves = append(s.leaves, h)
	return
}
func (s *InMemoryMerkleTreeStorage) storeInternal(h Hash, n kvMerkleTreeInternal) {
	s.nodes[h] = n
	return
}

func (s *InMemoryMerkleTreeStorage) storeParent(child Hash, parent Hash) {
	s.parent[child] = parent
	return
}

func (s *InMemoryMerkleTreeStorage) appendRoot(h Hash) {
	s.roots = append(s.roots, h)
	return
}

type KVMerkleTree struct {
	kvMerkleTreeStorage
	mh     MerkleHasher
}

func (m *KVMerkleTree) GetSubtreeSize(node Hash) int {
	n, ok := m.getInternal(node)
	if ok {
		return n.subtreeSize
	}
	_, ok = m.getLeaf(node)
	if ok {
		return 1
	}
	panic("unknown node")
}

func (m *KVMerkleTree) GetRoots() []Hash {
	n := m.getNumRoots()
	roots := []Hash{}
	for i := 0; i < n; i++ {
		roots = append(roots, m.getRoot(i))
	}
	return roots
}

func (m *KVMerkleTree) GetChildren(node Hash) []Hash {
	n, ok := m.getInternal(node)
	if !ok {
		panic("unknown node")
	}
	return n.children
}

func (m *KVMerkleTree) GetProof(node Hash) []Hash {
	_, yes := m.getLeaf(node)
	if !yes {
		panic("node is not a leaf")
	}
	proof := []Hash{}
	for {
		parent, there := m.getParent(node)
		if !there {
			break
		}
		pn, ok := m.getInternal(parent)
		if !ok {
			panic("unknown node")
		}
		proof = append(proof, pn.children...)
		node = parent
	}
	return proof
}

func (m *KVMerkleTree) IsLeaf(node Hash) bool {
	_, ok := m.getLeaf(node)
	return ok
}

func (m *KVMerkleTree) GetData(node Hash) []byte {
	n, ok := m.getLeaf(node)
	if !ok {
		panic("unknown node")
	}
	return n.data
}

func (m *KVMerkleTree) GetPrevSibling(node Hash) Hash {
	n, ok := m.getLeaf(node)
	if !ok {
		panic("unknown node")
	}
	if n.index > 0 {
		return m.getLeafHashByIndex(n.index-1)
	} else {
		return Hash{}
	}
}

func NewKVMerkleTree(s kvMerkleTreeStorage, data [][]byte, dim int) *KVMerkleTree {
	mh := NewSHA256Hasher(dim)
	m := &KVMerkleTree{
		kvMerkleTreeStorage: s,
		mh:     mh,
	}

	idx := 0
	for len(data) > 0 {
		// compute the size of the next tree
		size := 1
		for size*dim <= len(data) {
			size = size * dim
		}
		var nextHashes []Hash
		for i := 0; i < size; i++ {
			l := kvMerkleTreeLeaf{
				data:  data[i],
				index: idx,
			}
			h := m.mh.HashData(data[i])
			nextHashes = append(nextHashes, h)
			m.appendLeaf(h, l)
			idx++
		}
		for len(nextHashes) > 1 {
			var hashes []Hash // it is important that we allocate a new array because
			// internal nodes are referencing into nextHashes
			nb := len(nextHashes) / dim
			for i := 0; i < nb; i++ {
				n := kvMerkleTreeInternal{
					children:    nextHashes[i*dim : i*dim+dim],
					subtreeSize: size / nb,
				}
				h := m.mh.ComputeParent(nextHashes[i*dim : i*dim+dim])
				m.storeInternal(h, n)
				hashes = append(hashes, h)
				for j := 0; j < dim; j++ {
					m.storeParent(nextHashes[i*dim+j], h)
				}
			}
			nextHashes = hashes
		}
		// append the root
		m.appendRoot(nextHashes[0])
		data = data[size:]
	}
	return m
}

var test MerkleTree = &KVMerkleTree{}
