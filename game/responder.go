package game

type ResponderSession struct {
	tree MerkleTree
	ptr Hash
}

type ResponderMessage interface{}

type NextChildren struct {
	Hashes []Hash
}

type StateTransition struct {
	From interface{}
	FromProof interface{}
	To interface{}
}

func (s *ResponderSession) revealTransition(h Hash) StateTransition {
	fh := s.tree.GetPrevSibling(h)
	return StateTransition{s.tree.GetData(fh), s.tree.GetProof(fh), s.tree.GetData(h)}
}

func NewResponderSession(tree MerkleTree, from Hash) (*ResponderSession, ResponderMessage) {
	s := &ResponderSession{tree, from}
	if s.tree.IsLeaf(s.ptr) {
		return s, s.revealTransition(s.ptr)
	} else {
		return s, NextChildren{s.tree.GetChildren(from)}
	}
}

func (s *ResponderSession) Downward(req ChallengerMessage) ResponderMessage {
	if _, correct := req.(OpenNext); !correct {
		panic("unexpected challenge type")
	}

	idx := req.(OpenNext).Index
	s.ptr = s.tree.GetChildren(s.ptr)[idx]
	if s.tree.IsLeaf(s.ptr) {
		return s.revealTransition(s.ptr)
	} else {
		return NextChildren{s.tree.GetChildren(s.ptr)}
	}
}
