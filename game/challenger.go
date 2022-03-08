package game

type ChallengerSession struct {
	Tree MerkleTree
	ptr  Hash
	I    <-chan ResponderMessage
	O    chan<- ChallengerMessage
}

type ChallengerMessage interface{}

type OpenNext struct {
	Index int
}

type StartRoot struct {
	Index int
}

func (s *ChallengerSession) setStartPtr(r MountainRange) StartRoot {
	roots := s.Tree.GetRoots()
	theirIdx := 0
	for idx, root := range roots {
		if root == r.Roots[idx] {
			continue
		} else {
			s.ptr = roots[idx]
			theirIdx = idx
		}
	}
	// their subtree must be smaller than us by factor of dim^k
	for s.Tree.GetSubtreeSize(s.ptr) > r.Sizes[theirIdx] {
		s.ptr = s.Tree.GetChildren(s.ptr)[0]
	}
	return StartRoot{theirIdx}
}

func (s *ChallengerSession) Run() {
	defer close(s.O)
	// first wait for mountain range
	msg := <-s.I
	mr := msg.(MountainRange)
	rt := s.setStartPtr(mr)
	s.O <- rt

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
