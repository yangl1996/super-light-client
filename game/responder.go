package game

type ResponderSession struct {
	Tree MerkleTree
	ptr  Hash
	I    <-chan ChallengerMessage
	O    chan<- ResponderMessage
}

type ResponderMessage interface{}

type NextChildren struct {
	Hashes []Hash
}

type StateTransition struct {
	From      interface{} // from contains the prev state
	FromProof []Hash
	To        interface{} // to contains the current state, and the tx that causes the transition
}

type MountainRange struct {
	Roots []Hash
	Size  int
}

var zeroHash = Hash{}

func (s *ResponderSession) mountainRange() MountainRange {
	r := MountainRange{
		Roots: s.Tree.GetRoots(),
	}
	for _, rt := range r.Roots {
		r.Size += s.Tree.GetSubtreeSize(rt)
	}
	return r
}

func (s *ResponderSession) revealTransition(h Hash) StateTransition {
	fh := s.Tree.GetPrevSibling(h)
	if fh != zeroHash {
		return StateTransition{s.Tree.GetData(fh), s.Tree.GetProof(fh), s.Tree.GetData(h)}
	} else {
		return StateTransition{nil, nil, s.Tree.GetData(h)}
	}
}

func (s *ResponderSession) Run() {
	defer close(s.O)
	if s.Tree.IsLeaf(s.ptr) {
		s.O <- s.revealTransition(s.ptr)
		return
	} else {
		s.O <- NextChildren{s.Tree.GetChildren(s.ptr)}
	}
	for req := range s.I {
		if _, correct := req.(OpenNext); !correct {
			panic("unexpected challenge type")
		}
		idx := req.(OpenNext).Index
		s.ptr = s.Tree.GetChildren(s.ptr)[idx]
		if s.Tree.IsLeaf(s.ptr) {
			s.O <- s.revealTransition(s.ptr)
			return
		} else {
			s.O <- NextChildren{s.Tree.GetChildren(s.ptr)}
		}
	}
}
