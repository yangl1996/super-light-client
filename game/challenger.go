package game

type ChallengerSession struct {
	tree MerkleTree
	ptr Hash
}

func NewChallengerSession(tree MerkleTree, from Hash) *ChallengerSession {
	return &ChallengerSession{tree, from}
}

type ChallengerMessage interface{}

type OpenNext struct {
	Index int
}

type Terminate struct {}

func (s *ChallengerSession) Downward(resp ResponderMessage) ChallengerMessage {
	if s.tree.IsLeaf(s.ptr) {
		if _, correct := resp.(StateTransition); !correct {
			panic("unexpected response type from the responder")
		}

		// the verifier should now check the state transition
		return Terminate{}
	} else {
		// find the diff in the next level
		if _, correct := resp.(NextChildren); !correct {
			panic("unexpected response type from the responder")
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
				return OpenNext{i}
			}
		}
		panic("identical children in bisection game")
	}
}
