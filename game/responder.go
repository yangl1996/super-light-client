package game

var zeroHash = Hash{}

type Message interface{}

type OpenNext struct {
	Index int
}

type StartRoot struct {
	Index int
}

type NextChildren struct {
	Hashes []Hash
}

type StateTransition struct {
	From      []byte // from contains the prev state
	FromProof []Hash
	To        []byte // to contains the current state, and the tx that causes the transition
}

type MountainRange struct {
	Roots []Hash
	Sizes []int
}

type Session struct {
	Tree MerkleTree
	I    <-chan Message
	O    chan<- Message
	ptr  Hash
}

func DecideChallenger(m ...MountainRange) int {
	winner := 0
	winnerSize := 0
	for i, mr := range m {
		size := 0
		for _, sz := range mr.Sizes {
			size += sz
		}
		if size > winnerSize {
			winner = i
			winnerSize = size
		}
	}
	return winner
}

func (s *Session) Run() {
	mr := s.mountainRange()
	s.O <- mr
	msg := <-s.I
	switch m := msg.(type) {
	case MountainRange:
		s.runChallenger(m)
	case StartRoot:
		s.runResponder(m)
	default:
		panic("unexpected message type")
	}
}

func (s *Session) runResponder(sr StartRoot) {
	defer close(s.O)
	s.ptr = s.Tree.GetRoots()[sr.Index]

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

func (s *Session) runChallenger(mr MountainRange) {
	defer close(s.O)
	rt, needGame := s.setStartPtr(mr)
	if !needGame {
		panic("do not need a game")
	}
	s.O <- rt

	if s.Tree.IsLeaf(s.ptr) {
		return
	}
	for resp := range s.I {
		// find the diff in the next level
		if _, correct := resp.(NextChildren); !correct {
			panic("unexpected response type")
		}

		respHashes := resp.(NextChildren).Hashes
		ourHashes := s.Tree.GetChildren(s.ptr)
		if len(respHashes) != len(ourHashes) {
			panic("incompatible dimensions of merkle trees")
		}
		found := false
		for i := range ourHashes {
			if ourHashes[i] != respHashes[i] {
				// go downwards to the conflicting child
				s.ptr = ourHashes[i]
				s.O <- OpenNext{i}
				found = true
				break
			}
		}
		if !found {
			panic("identical children in bisection game")
		}
		if s.Tree.IsLeaf(s.ptr) {
			return
		}
	}
}

func (s *Session) mountainRange() MountainRange {
	r := MountainRange{
		Roots: s.Tree.GetRoots(),
	}
	for _, rt := range r.Roots {
		r.Sizes = append(r.Sizes, s.Tree.GetSubtreeSize(rt))
	}
	return r
}

func (s *Session) revealTransition(h Hash) StateTransition {
	fh := s.Tree.GetPrevSibling(h)
	if fh != zeroHash {
		return StateTransition{s.Tree.GetData(fh), s.Tree.GetProof(fh), s.Tree.GetData(h)}
	} else {
		return StateTransition{nil, nil, s.Tree.GetData(h)}
	}
}

func (s *Session) setStartPtr(r MountainRange) (StartRoot, bool) {
	roots := s.Tree.GetRoots()
	theirIdx := 0
	// look for the first root that is different
	for theirIdx, s.ptr = range roots {
		// compare subtree size
		if s.Tree.GetSubtreeSize(s.ptr) == r.Sizes[theirIdx] {
			// if subtree size is the same, then the root must be the same or we need
			// to look into it
			if s.ptr != r.Roots[theirIdx] {
				return StartRoot{theirIdx}, true
			}
		} else {
			// we found the first subtree with a different size
			break
		}
	}
	// scan until the end of peer's roots
	childIdx := 0
	for theirIdx < len(r.Roots) {
		// go down until our children has the same size as peer's subtree
		for s.Tree.GetSubtreeSize(s.Tree.GetChildren(s.ptr)[childIdx]) > r.Sizes[theirIdx] {
			s.ptr = s.Tree.GetChildren(s.ptr)[childIdx]
			childIdx = 0
		}
		if s.Tree.GetChildren(s.ptr)[childIdx] == r.Roots[theirIdx] {
			// go to sibling
			childIdx += 1
		} else {
			s.ptr = s.Tree.GetChildren(s.ptr)[childIdx]
			return StartRoot{theirIdx}, true
		}
		theirIdx += 1
	}
	return StartRoot{}, false // no need for bisection game
}
