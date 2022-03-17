package game

// Verifier implements the light client.
type Verifier struct {
	To   []chan<- Message
	From []<-chan Message

	Dim int
	MerkleHasher
}

const (
	BothWin = -1
)

// Match runs a match between a challenger and a prover. It takes the indices of the
// two parties, and the mountain range reported by the prover, which should have a
// shorter ledger than the challenger. It returns the index of the winner.
func (v *Verifier) Match(cidx, pidx int, pmr MountainRange) int {
	var diffIdx int
	var responderPtr Hash
	var responderSize int

	// send the mountain range to the challenger, and wait for it to pick the start point
	// for the game
	v.To[cidx] <- pmr
	var sr StartRoot
	switch m := (<-v.From[cidx]).(type) {
	case StartRoot:
		sr = m
	case NestedLedger:
		return BothWin
	default:
		return pidx	// unexpected type from the challenger
	}

	v.To[pidx] <- sr
	responderPtr = pmr.Roots[sr.Index]
	responderSize = pmr.Sizes[sr.Index]

	// run the bisection game to find the first disargeement
	for responderSize > 1 {
		// wait for the opening from the responder
		nc, ok := (<-v.From[pidx]).(NextChildren)
		if !ok {
			return cidx
		}
		if v.MerkleHasher.ComputeParent(nc.Hashes) != responderPtr {
			// responder loses because the opening does not match the parent hash
			return cidx
		}
		v.To[cidx] <- nc
		// wait for the index to open next
		on, ok := (<-v.From[cidx]).(OpenNext)
		if !ok {
			return pidx
		}
		v.To[pidx] <- on
		responderSize /= v.Dim
		responderPtr = nc.Hashes[on.Index]
		diffIdx = diffIdx*v.Dim + on.Index
	}
	var diffPrevTreeIdx int // the tree root idx of the leaf prev to the diff point (st.From below)
	if diffIdx == 0 {
		diffPrevTreeIdx = sr.Index - 1
	} else {
		diffPrevTreeIdx = sr.Index
	}
	// finish computing the diff index by adding the sizes the subtrees that are skipped
	{
		i := 0
		for i < sr.Index {
			diffIdx += pmr.Sizes[i]
			i += 1
		}
	}

	// wait for the responder to open the leaf
	st, ok := (<-v.From[pidx]).(StateTransition)
	if !ok {
		return cidx
	}
	// TODO: verify if st.To has index diffIdx and st.From has index diffIdx-1
	if v.MerkleHasher.HashData(st.To) != responderPtr {
		// incorrect hash of the opened leaf
		return cidx
	}
	if diffIdx != 0 {
		if !v.MerkleHasher.CheckProof(st.From, st.FromProof, pmr.Roots[diffPrevTreeIdx]) {
			// incorrect proof of the previous node
			return cidx
		}
	} else {
		if len(st.FromProof) != 0 || st.From != nil {
			// nonempty prev node when the diff is at index 0
			return cidx
		}
	}
	// TODO: verify the state transition
	return pidx
}

func (v *Verifier) Run() (MountainRange, int) {
	if len(v.To) != len(v.From) {
		panic("verifier launched with different incoming channels and outgoing channels")
	}

	// wait for everyone to send the mountain range
	// TODO: timeout
	mr := make([]MountainRange, len(v.From))
	for i := range mr {
		v.To[i] <- GetMountainRange{}
		mr[i] = (<-v.From[i]).(MountainRange)
		if len(mr[i].Sizes) != len(mr[i].Roots) {
			panic("different length of root and size array")
		}
		if len(mr[i].Sizes) > 1 {
			for j := 1; j < len(mr[i].Sizes); j++ {
				if mr[i].Sizes[j] > mr[i].Sizes[j-1] {
					panic("increasing size in size array")
				}
				if mr[i].Sizes[j-1]%mr[i].Sizes[j] != 0 {
					panic("noninteger size scale")
				}
				scale := mr[i].Sizes[j-1] / mr[i].Sizes[j]
				for scale != 1 {
					if scale%v.Dim != 0 {
						panic("scale not exponential of dimension")
					}
					scale = scale / v.Dim
				}
			}
		}
	}

	sizes := make([]int, len(v.From))
	for i := range sizes {
		for _, s := range mr[i].Sizes {
			sizes[i] += s
		}
	}
	safe := make(map[int]struct{})

	findLargestSafe := func() (int, int) {
		largestSafe := -1
		largestSafeSize := -1
		for k := range safe {
			if sizes[k] > largestSafeSize {
				largestSafe = k
				largestSafeSize = sizes[k]
			}
		}
		return largestSafe, largestSafeSize
	}

	for i := range sizes {
		if len(safe) == 0 {
			// no one is safe; the current peer automatically wins
			safe[i] = struct{}{}
		} else {
			// use the current peer to challenge the largest peer in the safe set
			// until all safe peer have lost, or the current peer has lost, or both win
			for {
				// use whoever that is larger to challenge the other
				largestSafe, largestSafeSize := findLargestSafe()
				var res int
				if largestSafeSize > sizes[i] {
					res = v.Match(largestSafe, i, mr[i])
				} else {
					res = v.Match(i, largestSafe, mr[largestSafe])
				}
				if res == BothWin {
					// both wins; the ledgers appear to be compatible
					safe[i] = struct{}{}
					break
				} else if res == i {
					// the new peer wins; remove the one in the safe set, and try to compete against the next
					delete(safe, largestSafe)
					if len(safe) == 0 {
						safe[i] = struct{}{}
						break
					}
				} else if res == largestSafe {
					// the one in the safe set wins; no point to continue
					break
				} else {
					panic("unreachable")
				}
			}
		}
	}

	winner, _ := findLargestSafe()
	return mr[winner], winner
}
