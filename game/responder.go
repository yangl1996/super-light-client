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
	Data interface{}
}

func NewResponderSession(tree MerkleTree, from Hash) (*ResponderSession, ResponderMessage) {
	s := &ResponderSession{tree, from}
	if s.tree.IsLeaf(s.ptr) {
		// TODO: the responder should prove state transition. We are now just returning the
		// leaf data.
		return s, StateTransition{s.tree.GetData(s.ptr)}
	} else {
		children := s.tree.GetChildren(from)
		return s, NextChildren{children}
	}
}

func (s *ResponderSession) Downward(idx int) ResponderMessage {
	s.ptr = s.tree.GetChildren(s.ptr)[idx]
	if s.tree.IsLeaf(s.ptr) {
		return StateTransition{s.tree.GetData(s.ptr)}
	} else {
		return NextChildren{s.tree.GetChildren(s.ptr)}
	}
}
