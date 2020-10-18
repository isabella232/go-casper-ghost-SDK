package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/prysmaticlabs/go-ssz"
)

func ProcessDeposits(state *core.State, deposits []*core.Deposit) error {
	for _, deposit := range deposits {
		if err := processDeposit(state, deposit); err != nil {
			return err
		}
	}
	return nil
}

/**
def process_deposit(state: BeaconState, deposit: Deposit) -> None:
    # Verify the Merkle branch
    assert is_valid_merkle_branch(
        leaf=hash_tree_root(deposit.data),
        branch=deposit.proof,
        depth=DEPOSIT_CONTRACT_TREE_DEPTH + 1,  # Add 1 for the List length mix-in
        index=state.eth1_deposit_index,
        root=state.eth1_data.deposit_root,
    )

    # Deposits must be processed in order
    state.eth1_deposit_index += 1

    pubkey = deposit.data.pubkey
    amount = deposit.data.amount
    validator_pubkeys = [v.pubkey for v in state.validators]
    if pubkey not in validator_pubkeys:
        # Verify the deposit signature (proof of possession) which is not checked by the deposit contract
        deposit_message = DepositMessage(
            pubkey=deposit.data.pubkey,
            withdrawal_credentials=deposit.data.withdrawal_credentials,
            amount=deposit.data.amount,
        )
        domain = compute_domain(DOMAIN_DEPOSIT)  # Fork-agnostic domain since deposits are valid across forks
        signing_root = compute_signing_root(deposit_message, domain)
        if not bls.Verify(pubkey, signing_root, deposit.data.signature):
            return

        # Add validator and balance entries
        state.validators.append(get_validator_from_deposit(state, deposit))
        state.balances.append(amount)
    else:
        # Increase balance by deposit amount
        index = ValidatorIndex(validator_pubkeys.index(pubkey))
        increase_balance(state, index, amount)
 */
func processDeposit(state *core.State, deposit *core.Deposit) error {
	// Verify the Merkle branch
	if err := verifyDeposit(state, deposit); err != nil {
		return fmt.Errorf("process deposit: %s", err.Error())
	}

	// Deposits must be processed in order
	state.Eth1DepositIndex ++

	pubkey := deposit.Data.PublicKey
	amount := deposit.Data.Amount
	bp := shared.BPByPubkey(state, pubkey)
	if bp == nil {
		// Verify the deposit signature (proof of possession) which is not checked by the deposit contract
		depositMsg := &core.DepositMessage{
			PublicKey:             deposit.Data.PublicKey,
			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
			Amount:                deposit.Data.Amount,
		}
		domain, err := shared.ComputeDomain(params.ChainConfig.DomainDeposit, nil, nil)
		if err != nil {
			return err
		}
		root, err := shared.ComputeSigningRoot(depositMsg, domain)
		if err != nil {
			return err
		}
		if res, err := shared.VerifySignature(root[:], deposit.Data.PublicKey, deposit.Data.Signature); !res || err != nil {
			if err != nil {
				return fmt.Errorf("process deposit: %s", err.Error())
			}
			return fmt.Errorf("process deposit: sig not verified")
		}

		// Add validator and balance entries
		state.BlockProducers = append(state.BlockProducers, GetBPFromDeposit(state, deposit))
	} else {
		// Increase balance by deposit amount
		shared.IncreaseBalance(state, bp.Id, amount)
	}

	return nil
}

func verifyDeposit(state *core.State, deposit *core.Deposit) error {
	// Verify Merkle proof of deposit and deposit trie root.
	if deposit == nil || deposit.Data == nil {
		return fmt.Errorf("received nil deposit or nil deposit data")
	}
	if state.Eth1Data == nil {
		return fmt.Errorf("received nil eth1data in the beacon state")
	}

	receiptRoot := state.Eth1Data.DepositRoot
	leaf, err := ssz.HashTreeRoot(deposit.Data)
	if err != nil {
		return fmt.Errorf("could not tree hash deposit data")
	}
	if ok := shared.VerifyMerkleBranch(
			receiptRoot,
			leaf[:],
			int(state.Eth1DepositIndex),
			deposit.Proof,
			params.ChainConfig.DepositContractTreeDepth,
		); !ok {
		return fmt.Errorf("deposit merkle branch of deposit root did not verify for root: %#x", receiptRoot)
	}
	return nil
}

/**
def get_validator_from_deposit(state: BeaconState, deposit: Deposit) -> Validator:
    amount = deposit.data.amount
    effective_balance = min(amount - amount % EFFECTIVE_BALANCE_INCREMENT, MAX_EFFECTIVE_BALANCE)

    return Validator(
        pubkey=deposit.data.pubkey,
        withdrawal_credentials=deposit.data.withdrawal_credentials,
        activation_eligibility_epoch=FAR_FUTURE_EPOCH,
        activation_epoch=FAR_FUTURE_EPOCH,
        exit_epoch=FAR_FUTURE_EPOCH,
        withdrawable_epoch=FAR_FUTURE_EPOCH,
        effective_balance=effective_balance,
    )
 */
func GetBPFromDeposit(state *core.State, deposit *core.Deposit) *core.BlockProducer {
	amount := deposit.Data.Amount
	effBalance := shared.Min(amount - amount % params.ChainConfig.EffectiveBalanceIncrement, params.ChainConfig.MaxEffectiveBalance)

	return &core.BlockProducer{
		Id:                         uint64(len(state.BlockProducers)),
		PubKey:                     deposit.Data.PublicKey,
		CDTBalance:                 0,
		EffectiveBalance:           effBalance,
		Balance:                    effBalance,
		Slashed:                    false,
		Active:                     true,
		ExitEpoch:                  params.ChainConfig.FarFutureEpoch,
		ActivationEpoch:            params.ChainConfig.FarFutureEpoch,
		ActivationEligibilityEpoch: params.ChainConfig.FarFutureEpoch,
		WithdrawableEpoch:          params.ChainConfig.FarFutureEpoch,
		WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
	}
}