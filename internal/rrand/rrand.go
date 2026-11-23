// Package rrand reproduces the random numbers of R's default generator, so that a Go program
// seeded like set.seed(seed) draws exactly what R's runif() and sample() draw.
//
// It follows R's own sources (src/main/RNG.c) for the three steps between set.seed and sample:
// the scrambling of the seed, the conversion of a 32-bit Mersenne Twister output to a number
// in (0, 1), and the rejection sampling of an index. It was checked against R 4.4.2 with
// RNGkind "Mersenne-Twister" and sample.kind "Rejection" (the defaults since R 3.6.0); the
// reference values are in testdata/r-reference.
//
// A Generator is not safe for concurrent use: give each goroutine its own.
package rrand

import "math/bits"

const (
	stateSize  = 624        // words of state in the 32-bit Mersenne Twister
	middleWord = 397        // offset of the word mixed into each new state word
	matrixA    = 0x9908b0df // the twist matrix, applied when the mixed word is odd
	upperMask  = 0x80000000 // most significant bit of a state word
	lowerMask  = 0x7fffffff // the other 31 bits

	seedMultiplier = 69069                  // R's linear congruential scrambler: seed*69069 + 1
	scrambleRounds = 50                     // rounds R applies before filling the state
	toUnitInterval = 2.3283064365386963e-10 // 2^-32, R's scale from a 32-bit output to [0, 1)
	bitsPerDraw    = 16                     // random bits R takes from each uniform number
)

// Generator is a Mersenne Twister in the state R would have after set.seed.
type Generator struct {
	state [stateSize]uint32
	next  int // position of the next unused state word; stateSize means "refill first"
}

// New returns a generator in the state R reaches after set.seed(seed).
func New(seed uint32) *Generator {
	for round := 0; round < scrambleRounds; round++ {
		seed = seedMultiplier*seed + 1
	}
	// R fills 625 words; the first is a position counter that it then overwrites, so the
	// value drawn for it is discarded here.
	seed = seedMultiplier*seed + 1
	generator := &Generator{next: stateSize}
	for word := range generator.state {
		seed = seedMultiplier*seed + 1
		generator.state[word] = seed
	}
	return generator
}

// twist combines the top bit of one state word with the low 31 bits of the next and applies
// the Mersenne Twister's recurrence matrix.
func twist(upper, lower uint32) uint32 {
	mixed := (upper & upperMask) | (lower & lowerMask)
	twisted := mixed >> 1
	if mixed&1 == 1 {
		twisted ^= matrixA
	}
	return twisted
}

// refill replaces all state words with the next 624. The recurrence reads the word
// middleWord places ahead, wrapping around; three plain loops cover the part before the
// wrap, the part after it, and the last word, so no index needs a remainder operation.
func (generator *Generator) refill() {
	state := &generator.state
	word := 0
	for ; word < stateSize-middleWord; word++ {
		state[word] = state[word+middleWord] ^ twist(state[word], state[word+1])
	}
	for ; word < stateSize-1; word++ {
		state[word] = state[word+middleWord-stateSize] ^ twist(state[word], state[word+1])
	}
	state[stateSize-1] = state[middleWord-1] ^ twist(state[stateSize-1], state[0])
	generator.next = 0
}

// uint32 returns the next tempered 32-bit output.
func (generator *Generator) uint32() uint32 {
	if generator.next >= stateSize {
		generator.refill()
	}
	output := generator.state[generator.next]
	generator.next++
	output ^= output >> 11
	output ^= (output << 7) & 0x9d2c5680
	output ^= (output << 15) & 0xefc60000
	output ^= output >> 18
	return output
}

// Float64 returns the next uniform number strictly inside (0, 1), as R's runif(1) does.
func (generator *Generator) Float64() float64 {
	value := float64(generator.uint32()) * toUnitInterval
	// R never returns exactly 0: it nudges it half a step inwards. (R guards 1 the same way,
	// but the largest 32-bit output times 2^-32 is already below 1, so no guard is needed.)
	if value <= 0 {
		return 0.5 * toUnitInterval
	}
	return value
}

// Index returns a uniformly distributed index in [0, n), drawn as R draws each element of
// sample.int(n, size, replace = TRUE) (R's result is this index plus one). n must be at
// least 1 and below 2^31, the range in which R uses this method.
//
// R builds a candidate from just enough random bits to cover n, 16 bits per uniform number,
// and draws again while the candidate is n or more. R takes those 16 bits as
// floor(unif_rand() * 65536); that is exactly the top 16 bits of the generator's 32-bit
// output, so they are taken directly, without a detour through floating point.
func (generator *Generator) Index(n int) int {
	// The number of bits needed to write n-1 equals ceil(log2(n)), R's bit count, and
	// costs no logarithm.
	bitCount := bits.Len(uint(n - 1))
	mask := uint64(1)<<bitCount - 1
	for {
		var candidate uint64
		for filled := 0; filled <= bitCount; filled += bitsPerDraw {
			candidate = candidate<<bitsPerDraw | uint64(generator.uint32()>>bitsPerDraw)
		}
		candidate &= mask
		if candidate < uint64(n) {
			return int(candidate)
		}
	}
}
