package shared

import "bytes"

// VerifyMerkleBranch verifies a Merkle branch against a root of a trie.
func VerifyMerkleBranch(root []byte, item []byte, merkleIndex int, proof [][]byte, depth uint64) bool {
	if len(proof) != int(depth)+1 {
		return false
	}
	node := SliceToByte32(item)
	for i := 0; i <= int(depth); i++ {
		if (uint64(merkleIndex) / PowerOf2(uint64(i)) % 2) != 0 {
			node = Hash(append(proof[i], node[:]...))
		} else {
			node = Hash(append(node[:], proof[i]...))
		}
	}

	return bytes.Equal(root, node[:])
}
