package shared

import (
	"github.com/minio/sha256-simd"
	"hash"
	"sync"
)

var sha256Pool = sync.Pool{New: func() interface{} {
	return sha256.New()
}}

// Hash defines a function that returns the sha256 checksum of the data passed in.
// https://github.com/ethereum/eth2.0-specs/blob/v0.9.3/specs/core/0_beacon-chain.md#hash
// Copied from prysm https://github.com/prysmaticlabs/prysm/blob/master/shared/hashutil/hash.go#L26-L47
func Hash(data []byte) [32]byte {
	h, ok := sha256Pool.Get().(hash.Hash)
	if !ok {
		h = sha256.New()
	}
	defer sha256Pool.Put(h)
	h.Reset()

	var b [32]byte

	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash

	// #nosec G104
	h.Write(data)
	h.Sum(b[:0])

	return b
}