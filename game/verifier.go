package game

type Verifier struct {
	toC chan<- ResponderMessage
	fromC <-chan ChallengerMessage

	toR chan<- ChallengerMessage
	fromR <-chan ResponderMessage

	mh MerkleHasher
}

func (v *Verifier) Run() {
	// the responder should send the first message
	for rm := range v.fromR {
		switch m := rm.(type) {
		case NextChildren:
			print(len(m.Hashes))	// use m so compiler does not complain
		}
	}
}
