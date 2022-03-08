package game

type ChallengerSession struct {
	tree MerkleTree
	ptr Hash
	I <-chan ResponderMessage
	O chan<- ChallengerMessage
}

type ChallengerMessage interface{}

type OpenNext struct {
	Index int
}

func (s *ChallengerSession) Run() {
	defer close(s.O)
	for resp := range s.I {
		// find the diff in the next level
		if _, correct := resp.(NextChildren); !correct {
			panic("unexpected response type")
		}

		respHashes := resp.(NextChildren).Hashes
		ourHashes := s.tree.GetChildren(s.ptr)
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
		if s.tree.IsLeaf(s.ptr) {
			return
		}
	}
}

