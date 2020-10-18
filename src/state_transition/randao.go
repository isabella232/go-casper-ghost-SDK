package state_transition

import (
	"encoding/binary"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/ulule/deepcopier"
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
func processRANDAO (state *core.State, block *core.PoolBlock) error {
	bp := shared.GetBlockProducer(state, block.Proposer)
	if bp == nil {
		return fmt.Errorf("could not find BP")
	}

	data, domain, err := RANDAOSigningData(state)
	if err != nil {
		return err
	}
	res, err := shared.VerifyRandaoRevealSignature(data, domain, bp.PubKey, block.Body.RandaoReveal)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("randao sig not verified")
	}

	return processRANDAONoVerify(state, block)
}

// ProcessRandaoNoVerify generates a new randao mix to update
// in the beacon state's latest randao mixes slice.
//
// Spec pseudocode definition:
//     # Mix it in
//     state.latest_randao_mixes[get_current_epoch(state) % LATEST_RANDAO_MIXES_LENGTH] = (
//         xor(get_randao_mix(state, get_current_epoch(state)),
//             hash(body.randao_reveal))
//     )
func processRANDAONoVerify(state *core.State, block *core.PoolBlock) error {
	latestMix := make([]byte, 32)
	deepcopier.Copy(shared.GetLatestRandaoMix(state)).To(latestMix)
	hash := shared.Hash(block.Body.RandaoReveal)

	if len(hash) != len(latestMix) {
		return fmt.Errorf("randao reveal length doesn't match existing mix")
	}

	for i,x := range hash {
		latestMix[i] ^= x
	}

	state.Randao = append(state.Randao, &core.SlotAndBytes{
		Slot:                state.CurrentSlot,
		Bytes:               latestMix,
	})
	return nil
}

func RANDAOSigningData(state *core.State) (data []byte, domain []byte, err error)  {
	epoch := shared.GetCurrentEpoch(state)
	data = make([]byte, 8) // 64 bit
	binary.LittleEndian.PutUint64(data, epoch)

	domain, err = shared.GetDomain(state, params.ChainConfig.DomainRandao, epoch)
	if err != nil {
		return nil,nil, err
	}

	return data, domain, nil
}