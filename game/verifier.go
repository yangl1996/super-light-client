package game

type Verifier struct {
	To []chan<- Message
	From []<-chan Message

	Dim int
	MerkleHasher
}

func (v *Verifier) Run() StateTransition {
	var responderPtr Hash
	var responderSize int
	diffFirst := true // if the expected diff is the first element

	// wait for both to send mountain range
	// TODO: check the size of the subtrees
	mr := make([]MountainRange, 2)
	mr[0] = (<-v.From[0]).(MountainRange)
	mr[1] = (<-v.From[1]).(MountainRange)
	v.To[0] <- mr[1]
	v.To[1] <- mr[0]
	cidx := DecideChallenger(mr...)
	ridx := 1-cidx;

	// wait for challenger to set initial pointer
	sr := (<-v.From[cidx]).(StartRoot)
	responderPtr = mr[ridx].Roots[sr.Index]
	responderSize = mr[ridx].Sizes[sr.Index]
	v.To[ridx] <- sr
	if sr.Index != 0 {
		diffFirst = false
	}

	for responderSize > 1 {
		nc := (<-v.From[ridx]).(NextChildren)
		if v.MerkleHasher.ComputeParent(nc.Hashes) != responderPtr {
			panic("incorrect children pointers")
		}
		v.To[cidx] <- nc
		on := (<-v.From[cidx]).(OpenNext)
		responderSize /= v.Dim
		responderPtr = nc.Hashes[on.Index]
		v.To[ridx] <- on
		if on.Index != 0 {
			diffFirst = false
		}
	}

	st := (<-v.From[ridx]).(StateTransition)
	// TODO: we do not verify ordering. An index must go with the data in a real app
	if v.MerkleHasher.HashData(st.To) != responderPtr {
		panic("incorrect to data")
	}
	if !diffFirst {
		if !v.MerkleHasher.CheckProof(st.From, st.FromProof, mr[ridx].Roots...) {
			panic("incorrect from data")
		}
	} else {
		if len(st.FromProof) != 0 || st.From != nil {
			panic("nonempty from node when diff at 0")
		}
	}
	return st
}
