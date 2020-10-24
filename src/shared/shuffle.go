package shared

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const seedSize = int8(32)
const roundSize = int8(1)
const positionWindowSize = int8(4)
const pivotViewSize = seedSize + roundSize
const totalSize = seedSize + roundSize + positionWindowSize

var maxShuffleListSize uint64 = 1 << 31

// computeShuffledIndex returns the shuffled validator index corresponding to seed and index count.
// Spec pseudocode definition:
//   def compute_shuffled_index(index: ValidatorIndex, index_count: uint64, seed: Hash) -> ValidatorIndex:
//    """
//    Return the shuffled validator index corresponding to ``seed`` (and ``index_count``).
//    """
//    assert index < index_count
//
//    # Swap or not (https://link.springer.com/content/pdf/10.1007%2F978-3-642-32009-5_1.pdf)
//    # See the 'generalized domain' algorithm on page 3
//    for current_round in range(SHUFFLE_ROUND_COUNT):
//        pivot = bytes_to_int(hash(seed + int_to_bytes(current_round, length=1))[0:8]) % index_count
//        flip = ValidatorIndex((pivot + index_count - index) % index_count)
//        position = max(index, flip)
//        source = hash(seed + int_to_bytes(current_round, length=1) + int_to_bytes(position // 256, length=4))
//        byte = source[(position % 256) // 8]
//        bit = (byte >> (position % 8)) % 2
//        index = flip if bit else index
//
//    return ValidatorIndex(index)
func computeShuffledIndex(index uint64, indexCount uint64, seed [32]byte, shuffle bool, shuffleRoundCount uint64) (uint64, error) {
	if index >= indexCount {
		return 0, fmt.Errorf("input index %d out of bounds: %d",
			index, indexCount)
	}
	if indexCount > maxShuffleListSize {
		return 0, fmt.Errorf("list size %d out of bounds",
			indexCount)
	}
	rounds := uint8(shuffleRoundCount)
	round := uint8(0)
	if !shuffle {
		// Starting last round and iterating through the rounds in reverse, un-swaps everything,
		// effectively un-shuffling the list.
		round = rounds - 1
	}
	buf := make([]byte, totalSize, totalSize)
	posBuffer := make([]byte, 8, 8)
	hashfunc := sha256.Sum256

	// seed is always the first 32 bytes of the hash input, we never have to change this part of the buffer.
	copy(buf[:32], seed[:])
	for {
		buf[seedSize] = round
		hash := hashfunc(buf[:pivotViewSize])
		hash8 := hash[:8]
		hash8Int := fromBytes8(hash8)
		pivot := hash8Int % indexCount
		flip := (pivot + indexCount - index) % indexCount
		// Consider every pair only once by picking the highest pair index to retrieve randomness.
		position := index
		if flip > position {
			position = flip
		}
		// Add position except its last byte to []buf for randomness,
		// it will be used later to select a bit from the resulting hash.
		binary.LittleEndian.PutUint64(posBuffer[:8], position>>8)
		copy(buf[pivotViewSize:], posBuffer[:4])
		source := hashfunc(buf)
		// Effectively keep the first 5 bits of the byte value of the position,
		// and use it to retrieve one of the 32 (= 2^5) bytes of the hash.
		byteV := source[(position&0xff)>>3]
		// Using the last 3 bits of the position-byte, determine which bit to get from the hash-byte (note: 8 bits = 2^3)
		bitV := (byteV >> (position & 0x7)) & 0x1
		// index = flip if bit else index
		if bitV == 1 {
			index = flip
		}
		if shuffle {
			round++
			if round == rounds {
				break
			}
		} else {
			if round == 0 {
				break
			}
			round--
		}
	}
	return index, nil
}

// swapOrNot describes the main algorithm behind the shuffle where we swap bytes in the inputted value
// depending on if the conditions are met.
func swapOrNot(buf []byte, byteV byte, i uint64, input []uint64,
	j uint64, source [32]byte, hashFunc func([]byte) [32]byte) (byte, [32]byte) {
	if j&0xff == 0xff {
		// just overwrite the last part of the buffer, reuse the start (seed, round)
		binary.LittleEndian.PutUint32(buf[pivotViewSize:], uint32(j>>8))
		source = hashFunc(buf)
	}
	if j&0x7 == 0x7 {
		byteV = source[(j&0xff)>>3]
	}
	bitV := (byteV >> (j & 0x7)) & 0x1

	if bitV == 1 {
		input[i], input[j] = input[j], input[i]
	}
	return byteV, source
}

// fromBytes8 returns an integer which is stored in the little-endian format(8, 'little')
// from a byte array.
func fromBytes8(x []byte) uint64 {
	return binary.LittleEndian.Uint64(x)
}
