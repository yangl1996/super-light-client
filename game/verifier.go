package game

type Verifier struct {
	ToC   chan<- ResponderMessage
	FromC <-chan ChallengerMessage

	ToR   chan<- ChallengerMessage
	FromR <-chan ResponderMessage

	Dim int
	MerkleHasher
}

func (v *Verifier) Run() StateTransition {
	var responderPtr Hash
	var responderSize int
	diffFirst := true // if the expected diff is the first element
	// wait for responder to send mountain range
	mr := (<-v.FromR).(MountainRange)
	// TODO: check the size of the subtrees
	v.ToC <- mr

	// wait for challenger to set initial pointer
	sr := (<-v.FromC).(StartRoot)
	responderPtr = mr.Roots[sr.Index]
	responderSize = mr.Sizes[sr.Index]
	v.ToR <- sr
	if sr.Index != 0 {
		diffFirst = false
	}

	for responderSize > 1 {
		nc := (<-v.FromR).(NextChildren)
		if v.MerkleHasher.ComputeParent(nc.Hashes) != responderPtr {
			panic("incorrect children pointers")
		}
		v.ToC <- nc
		on := (<-v.FromC).(OpenNext)
		responderSize /= v.Dim
		responderPtr = nc.Hashes[on.Index]
		v.ToR <- on
		if on.Index != 0 {
			diffFirst = false
		}
	}

	st := (<-v.FromR).(StateTransition)
	// TODO: we do not verify ordering. An index must go with the data in a real app
	if v.MerkleHasher.HashData(st.To) != responderPtr {
		panic("incorrect to data")
	}
	if !diffFirst {
		if !v.MerkleHasher.CheckProof(st.From, st.FromProof, mr.Roots...) {
			panic("incorrect from data")
		}
	} else {
		if len(st.FromProof) != 0 || st.From != nil {
			panic("nonempty from node when diff at 0")
		}
	}
	return st
}
