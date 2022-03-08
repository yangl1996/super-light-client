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

type Terminate struct {}

func (s *ChallengerSession) Run() {
	defer close(s.O)
	for resp := range s.I {
		if s.tree.IsLeaf(s.ptr) {
			if _, correct := resp.(Terminate); !correct {
				panic("unexpected response type")
			}
			// the verifier should now check the state transition, know that the challenger
			// agrees that we have found the leaf
			s.O <- Terminate{}
			return
		} else {
			// find the diff in the next level
			if _, correct := resp.(NextChildren); !correct {
				panic("unexpected response type")
			}

			respHashes := resp.(NextChildren).Hashes
			ourHashes := s.tree.GetChildren(s.ptr)
			if len(respHashes) != len(ourHashes) {
				panic("incompatible dimensions of merkle trees")
			}
			for i := range ourHashes {
				if ourHashes[i] != respHashes[i] {
					// go downwards to the conflicting child
					s.ptr = ourHashes[i]
					s.O <- OpenNext{i}
				}
			}
			panic("identical children in bisection game")
		}
	}
}

