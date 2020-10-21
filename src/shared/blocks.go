package shared

import (
	"encoding/binary"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/htrutils"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
)

// BlockHeaderRoot computes the HashTreeRoot Merkleization of
// a BeaconBlockHeader struct according to the eth2
// Simple Serialize specification.
func BlockHeaderRoot(header *core.BlockHeader) ([32]byte, error) {
	fieldRoots := make([][]byte, 5)
	if header != nil {
		headerSlotBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(headerSlotBuf, header.Slot)
		headerSlotRoot := bytesutil.ToBytes32(headerSlotBuf)
		fieldRoots[0] = headerSlotRoot[:]
		proposerIdxBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(proposerIdxBuf, header.ProposerIndex)
		proposerIndexRoot := bytesutil.ToBytes32(proposerIdxBuf)
		fieldRoots[1] = proposerIndexRoot[:]
		parentRoot := bytesutil.ToBytes32(header.ParentRoot)
		fieldRoots[2] = parentRoot[:]
		stateRoot := bytesutil.ToBytes32(header.StateRoot)
		fieldRoots[3] = stateRoot[:]
		bodyRoot := bytesutil.ToBytes32(header.BodyRoot)
		fieldRoots[4] = bodyRoot[:]
	}
	return htrutils.BitwiseMerkleize(hashutil.CustomSHA256Hasher(), fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
}
