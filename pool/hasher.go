package pool

import (
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"sync"
)

// hasherPool manages a pool of SHA3-256 hashers.
var hasherPool = sync.Pool{
	New: func() interface{} {
		return sha3.NewLegacyKeccak256()
	},
}

// getHasher retrieves a hasher from the pool.
func GetHasher() crypto.KeccakState {
	return hasherPool.Get().(crypto.KeccakState)
}

// putHasher resets the hasher and returns it to the pool.
func PutHasher(h crypto.KeccakState) {
	h.Reset()
	hasherPool.Put(h)
}

// hashData computes the SHA3-256 hash of the input data using a pooled hasher.
func HashData(data []byte) []byte {
	hasher := GetHasher()
	defer PutHasher(hasher) // Return the hasher to the pool after use

	hasher.Write(data)
	return hasher.Sum(nil)
}
