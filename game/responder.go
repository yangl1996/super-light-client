package game

type ResponderSession struct {
	tree MerkleTree
	ptr Hash
	I <-chan ChallengerMessage
	O chan<- ResponderMessage
}

type ResponderMessage interface{}

type NextChildren struct {
	Hashes []Hash
}

type StateTransition struct {
	From interface{}	// from contains the prev state
	FromProof interface{}
	To interface{}		// to contains the current state, and the tx that causes the transition
}

func (s *ResponderSession) revealTransition(h Hash) StateTransition {
	fh := s.tree.GetPrevSibling(h)
	return StateTransition{s.tree.GetData(fh), s.tree.GetProof(fh), s.tree.GetData(h)}
}

func (s *ResponderSession) Run() {
	defer close(s.O)
	if s.tree.IsLeaf(s.ptr) {
		s.O <- s.revealTransition(s.ptr)
		return
	} else {
		s.O <- NextChildren{s.tree.GetChildren(s.ptr)}
	}
	for req := range s.I {
		if _, correct := req.(OpenNext); !correct {
			panic("unexpected challenge type")
		}
		idx := req.(OpenNext).Index
		s.ptr = s.tree.GetChildren(s.ptr)[idx]
		if s.tree.IsLeaf(s.ptr) {
			s.O <- s.revealTransition(s.ptr)
			return
		} else {
			s.O <- NextChildren{s.tree.GetChildren(s.ptr)}
		}
	}
}

