package state_transition

import (
	"encoding/binary"
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
)

// Spec pseudocode definition:
//   def process_randao(state: BeaconState, body: BeaconBlockBody) -> None:
//    epoch = get_current_epoch(state)
//    # Verify RANDAO reveal
//    proposer = state.validators[get_beacon_proposer_index(state)]
//    signing_root = compute_signing_root(epoch, get_domain(state, DOMAIN_RANDAO))
//    assert bls.Verify(proposer.pubkey, signing_root, body.randao_reveal)
//    # Mix in RANDAO reveal
//    mix = xor(get_randao_mix(state, epoch), hash(body.randao_reveal))
//    state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR] = mix
func processRANDAO (state *core.State, block *core.Block) error {
	validator := shared.GetValidator(state, block.Proposer)
	if validator == nil {
		return fmt.Errorf("could not find BP")
	}

	epochByts, domain, err := RANDAOSigningData(state)
	if err != nil {
		return err
	}
	res, err := shared.VerifyRandaoRevealSignature(epochByts, domain, validator.PublicKey, block.Body.RandaoReveal)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("randao sig not verified")
	}

	return processRANDAONoVerify(state, block)
}

/**
# Mix in RANDAO reveal
    mix = xor(get_randao_mix(state, epoch), hash(body.randao_reveal))
    state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR] = mix
 */
func processRANDAONoVerify(state *core.State, block *core.Block) error {
	latestMix := make([]byte, 32)
	copy(latestMix, shared.GetRandaoMix(state, shared.GetCurrentEpoch(state)))
	hash := hashutil.Hash(block.Body.RandaoReveal)

	if len(hash) != len(latestMix) {
		return fmt.Errorf("randao reveal length doesn't match existing mix")
	}

	for i,x := range hash {
		latestMix[i] ^= x
	}

	state.RandaoMixes[shared.GetCurrentEpoch(state) % params.ChainConfig.EpochsPerHistoricalVector] = latestMix
	return nil
}

func RANDAOSigningData(state *core.State) ([32]byte, []byte, error)  {
	epoch := shared.GetCurrentEpoch(state)
	epochByts := make([]byte, 32) // 64 bit
	binary.LittleEndian.PutUint64(epochByts, epoch)

	domain, err := shared.GetDomain(state, params.ChainConfig.DomainRandao, epoch)
	if err != nil {
		return [32]byte{},nil, err
	}

	return bytesutil.ToBytes32(epochByts), domain, nil
}