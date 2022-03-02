package game

type MerkleMountainRange struct {
	Trees []MerkleTree
	Sizes []int	// number of elements
}

func (m *MerkleMountainRange) GetPeaks() []Hash {
	res := []Hash{}
	for _, t := range m.Trees {
		res = append(res, t.GetRoot())
	}
	return res
}

type MerkleTree interface {
	GetRoot() Hash
    GetChildren(node Hash) []Hash
    GetProof(node Hash) []Hash
	IsLeaf(node Hash) bool
	GetData(node Hash) interface{}
}
