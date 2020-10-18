package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

func ProcessExits(state *core.State, exits []*core.SignedVoluntaryExit) error {
	for _, exit := range exits {
		if err := processExit(state, exit); err != nil {
			return err
		}
	}
	return nil
}

/**
def process_voluntary_exit(state: BeaconState, signed_voluntary_exit: SignedVoluntaryExit) -> None:
    voluntary_exit = signed_voluntary_exit.message
    validator = state.validators[voluntary_exit.validator_index]
    # Verify the validator is active
    assert is_active_validator(validator, get_current_epoch(state))
    # Verify exit has not been initiated
    assert validator.exit_epoch == FAR_FUTURE_EPOCH
    # Exits must specify an epoch when they become valid; they are not valid before then
    assert get_current_epoch(state) >= voluntary_exit.epoch
    # Verify the validator has been active long enough
    assert get_current_epoch(state) >= validator.activation_epoch + SHARD_COMMITTEE_PERIOD
    # Verify signature
    domain = get_domain(state, DOMAIN_VOLUNTARY_EXIT, voluntary_exit.epoch)
    signing_root = compute_signing_root(voluntary_exit, domain)
    assert bls.Verify(validator.pubkey, signing_root, signed_voluntary_exit.signature)
    # Initiate exit
    initiate_validator_exit(state, voluntary_exit.validator_index)
 */
func processExit(state *core.State, exit *core.SignedVoluntaryExit) error {
	voluntaryExit := exit.Exit
	bp := shared.GetBlockProducer(state, voluntaryExit.ValidatorIndex)
	if bp == nil {
		return fmt.Errorf("process exit: BP %d not found", voluntaryExit.ValidatorIndex)
	}

	// Verify the validator is active
	if !shared.IsActiveBP(bp, shared.GetCurrentEpoch(state)) {
		return fmt.Errorf("process exit: BP %d not active", voluntaryExit.ValidatorIndex)
	}
	// Verify exit has not been initiated
	if bp.ExitEpoch != params.ChainConfig.FarFutureEpoch {
		return fmt.Errorf("process exit: BP %d has started exit", voluntaryExit.ValidatorIndex)
	}
	// Exits must specify an epoch when they become valid; they are not valid before then
	if shared.GetCurrentEpoch(state) < voluntaryExit.Epoch {
		return fmt.Errorf("process exit: Exits must specify an epoch when they become valid; they are not valid before then")
	}
	// Verify the validator has been active long enough
	if shared.GetCurrentEpoch(state) < bp.ActivationEpoch + params.ChainConfig.ShardCommitteePeriod {
		return fmt.Errorf("process exit: Verify the validator has been active long enough")
	}
	// Verify signature
	domain, err := shared.GetDomain(state, params.ChainConfig.DomainVoluntaryExit, voluntaryExit.Epoch)
	if err != nil {
		return fmt.Errorf("process exit: %s", err.Error())
	}
	root, err := shared.ComputeSigningRoot(voluntaryExit, domain)
	if err != nil {
		return fmt.Errorf("process exit: %s", err.Error())
	}
	if res, err := shared.VerifySignature(root[:], bp.PubKey, exit.Signature); !res || err != nil {
		if err != nil {
			return fmt.Errorf("process exit: %s", err.Error())
		}
		return fmt.Errorf("process exit: sig invalid")
	}

	shared.InitiateBlockProducerExit(state, voluntaryExit.ValidatorIndex)

	return nil
}
