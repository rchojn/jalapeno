package engine

import (
	"math/big"
	"math/rand"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/gofrs/uuid"
)

var prngs = make(map[uuid.UUID]*rand.Rand)

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add additional template functions here
	f["stableRandomAlphanumeric"] = stableRandomAlphanumeric

	return f
}

const lowerAlpha = "abcdefghijlkmnopqrstuvwxyz"
const number = "0123456789"
const alphanumeric = lowerAlpha + number

// Generate a pseudo-random alphanumeric string of the given length
// such that the sequence of strings generated by successive calls
// is the same for a given anchor.
func stableRandomAlphanumeric(length int, anchor uuid.UUID) string {
	p := prngForAnchor(anchor)
	chars := make([]byte, length)
	for i := 0; i < length; i++ {
		chars[i] = alphanumeric[p.Intn(len(alphanumeric))]
	}
	return string(chars)
}

// Get the PRNG for the given anchor, by either seeding a new one
// or returning a previously seeded one.
func prngForAnchor(anchor uuid.UUID) *rand.Rand {
	p, ok := prngs[anchor]
	if !ok {
		i := big.NewInt(0)
		i.SetString(strings.ReplaceAll(anchor.String(), "-", ""), 16)
		seed := big.NewInt(0)
		// We throw away half of the bits of the UUID here with the Rsh
		// but probably it's still ok?
		p = rand.New(rand.NewSource(seed.Rsh(i, 1).Int64()))
		prngs[anchor] = p
	}
	return p
}

// Reset generated PRGNs.
func resetRngs() {
	prngs = make(map[uuid.UUID]*rand.Rand)
}
