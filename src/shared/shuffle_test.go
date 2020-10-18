package shared

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)


func getSeed(str string) [32]byte {
	var ret [32]byte
	_seed, _ := hex.DecodeString(str)
	copy(ret[:],_seed)

	return ret
}

func TestDeterministicShuffling(t *testing.T) {
	tests := []struct {
		testName string
		seed [32]byte
		rounds uint8
		indexes []uint64
		expected []uint64
	} {
		{
			testName:"shuffle, 5 rounds",
			seed: getSeed("b581262ce281d1e9deaf2f0158d7cd05217f1196d95956c5f55d837ccc3c8a9"),
			rounds: 5,
			indexes: []uint64{1,2,3,4},
			expected: []uint64{4,2,3,1},
		},
		{
			testName:"shuffle, 10 rounds",
			seed: getSeed("b581262ce281d1e9deaf2f0158d7cd05217f1196d95956c5f55d837ccc3c8a9"),
			rounds: 10,
			indexes: []uint64{1,2,3,4},
			expected: []uint64{1,2,3,4},
		},
		{
			testName:"shuffle, 15 rounds",
			seed: getSeed("b581262ce281d1e9deaf2f0158d7cd05217f1196d95956c5f55d837ccc3c8a9"),
			rounds: 15,
			indexes: []uint64{1,2,3,4},
			expected: []uint64{4,2,1,3},
		},
		{
			testName:"shuffle seed #2, 5 rounds",
			seed: getSeed("f536fd5464af265f824e9a62144e69ecc5ef0749e5be6743dd69e28b2362e6c4"),
			rounds: 5,
			indexes: []uint64{1,2,3,4},
			expected: []uint64{2,3,1,4},
		},
		{
			testName:"shuffle seed #2, 6 rounds",
			seed: getSeed("f536fd5464af265f824e9a62144e69ecc5ef0749e5be6743dd69e28b2362e6c4"),
			rounds: 6,
			indexes: []uint64{1,2,3,4},
			expected: []uint64{3,2,1,4},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			res, err := ShuffleList(test.indexes, test.seed, test.rounds)
			require.NoError(t,err)
			require.EqualValues(t, test.expected, res)
		})
	}
}
