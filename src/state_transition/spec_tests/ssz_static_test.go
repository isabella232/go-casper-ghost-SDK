package spec_tests

import (
	"encoding/hex"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	ssz "github.com/ferranbt/fastssz"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var nameToObject = map[string]func() interface{} {
	"AggregateAndProof": nil,
	"Attestation": func() interface{} { return new(core.Attestation) },
	"attestation": func() interface{} { return new(core.Attestation) },
	"AttestationData": func() interface{} { return new(core.AttestationData) },
	"AttesterSlashing": func() interface{} { return new(core.AttesterSlashing) },
	"attester_slashing": func() interface{} { return new(core.AttesterSlashing) },
	"BeaconBlock": func() interface{} { return new(core.Block) },
	"block": func() interface{} { return new(core.Block) },
	"BeaconBlockBody": func() interface{} { return new(core.BlockBody) },
	"BeaconBlockHeader": func() interface{} { return new(core.BlockHeader) },
	"block_header": func() interface{} { return new(core.BlockHeader) },
	"BeaconState": func() interface{} { return new(core.State) },
	"Checkpoint": func() interface{} { return new(core.Checkpoint) },
	"Deposit": func() interface{} { return new(core.Deposit) },
	"deposit": func() interface{} { return new(core.Deposit) },
	"DepositData": func() interface{} { return new(core.Deposit_DepositData) },
	"DepositMessage": func() interface{} { return new(core.DepositMessage) },
	"Eth1Block": nil,
	"Eth1Data": func() interface{} { return new(core.ETH1Data) },
	"Fork": func() interface{} { return new(core.Fork) },
	"ForkData": func() interface{} { return new(core.ForkData) },
	"HistoricalBatch": func() interface{} { return new(core.HistoricalBatch) },
	"IndexedAttestation": func() interface{} { return new(core.IndexedAttestation) },
	"PendingAttestation": func() interface{} { return new(core.PendingAttestation) },
	"ProposerSlashing": func() interface{} { return new(core.ProposerSlashing) },
	"proposer_slashing": func() interface{} { return new(core.ProposerSlashing) },
	"SignedAggregateAndProof": nil,
	"SignedBeaconBlock": func() interface{} { return new(core.SignedBlock) },
	"SignedBeaconBlockHeader": func() interface{} { return new(core.SignedBlockHeader) },
	"SignedVoluntaryExit": func() interface{} { return new(core.SignedVoluntaryExit) },
	"SigningData": nil,
	"Validator": func() interface{} { return new(core.Validator) },
	"VoluntaryExit": func() interface{} { return new(core.VoluntaryExit) },
	"voluntary_exit": func() interface{} { return new(core.SignedVoluntaryExit) },
}

func TestSpecSSZStaticMainnet(t *testing.T) {
	base, err := os.Getwd()
	require.NoError(t, err)

	type SSZRoots struct {
		Root        string `json:"root"`
		SigningRoot string `json:"signing_root"`
	}

	root := path.Join(base, rootSpecTestsFolder, "mainnet/phase0/ssz_static")
	files, err := ioutil.ReadDir(root)
	require.NoError(t, err)
	for _, file := range files {
		if file.IsDir() {
			t.Run(file.Name(), func(tt *testing.T) {
				if objFunc, ok := nameToObject[file.Name()]; ok && objFunc != nil {
					subDirsPath := path.Join(root, file.Name(), "ssz_random")
					subDirs, err := ioutil.ReadDir(subDirsPath)
					require.NoError(t, err)
					for _, subDir := range subDirs {
						tt.Run(subDir.Name(), func(ttt *testing.T) {
							obj := objFunc()

							// unmarshal SSZ to dedicated object
							byts, err := ioutil.ReadFile(path.Join(subDirsPath, subDir.Name(), "serialized.ssz"))
							require.NoError(ttt, err)
							require.NoError(ttt, obj.(ssz.Unmarshaler).UnmarshalSSZ(byts))

							// unmarshal expected root
							rootByts, err := ioutil.ReadFile(path.Join(subDirsPath, subDir.Name(), "roots.yaml"))
							require.NoError(ttt, err)
							root := &SSZRoots{}
							require.NoError(ttt, testutil.UnmarshalYaml(rootByts, root))
							expectedRoot, err := hex.DecodeString(root.Root[2:])
							require.NoError(ttt, err)

							// hash to root and compare
							targetRoot, err := obj.(ssz.HashRoot).HashTreeRoot()
							require.NoError(ttt, err)
							require.EqualValues(ttt, expectedRoot, targetRoot[:])
						})
					}
				} else {
					tt.Skip("no object function")
				}
			})
		}
	}
}