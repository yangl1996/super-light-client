package game

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
