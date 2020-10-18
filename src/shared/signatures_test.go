package shared

import (
	"encoding/hex"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/stretchr/testify/require"
	"testing"
)

func toByte(str string) []byte {
	ret, _ := hex.DecodeString(str)
	return ret
}

func TestBlockSigningRootOk(t *testing.T) {
	block := &core.PoolBlock{
		Slot:                 0,
		Proposer:             0,
		ParentRoot:           []byte{},
		StateRoot:            []byte{},
		Body:                 &core.PoolBlockBody{
			RandaoReveal:         []byte{},
			Attestations:         []*core.Attestation{},
			NewPoolReq:           []*core.CreateNewPoolRequest{},
		},
	}

	root, err := BlockSigningRoot(block, []byte("domain"))
	require.NoError(t, err)
	require.EqualValues(t, toByte("fe909cde8ee3253d2951cec3dfcee82b3b6613fadf0b08030eb038c309bf6ea9"),root[:])
}

func TestComputeDomainOk(t *testing.T) {
	tests := []struct{
		epoch uint64
		domainType []byte
		domain []byte
	}{
		{
			epoch: 1, domainType: []byte{0,0,0,0}, domain:toByte("0000000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			epoch: 1, domainType: []byte{1,0,0,0}, domain:toByte("0100000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			epoch: 1, domainType: []byte{2,0,0,0}, domain:toByte("0200000000000000000000000000000000000000000000000000000000000000"),
		},
	}

	for _, tt := range tests {
		d := ComputeDomain(tt.domainType, nil, nil)
		require.EqualValues(t, tt.domain, d)
	}
}

func TestDomainOk(t *testing.T) {
	tests := []struct{
		epoch uint64
		domainType []byte
		domain []byte
	}{
		{
			epoch: 1, domainType: []byte{0,0,0,0}, domain:toByte("0000000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			epoch: 1, domainType: []byte{1,0,0,0}, domain:toByte("0100000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			epoch: 1, domainType: []byte{2,0,0,0}, domain:toByte("0200000000000000000000000000000000000000000000000000000000000000"),
		},
	}

	for _, tt := range tests {
		d, err := GetDomain(tt.epoch, tt.domainType, tt.epoch)
		require.NoError(t, err)
		require.EqualValues(t, tt.domain, d)
	}
}