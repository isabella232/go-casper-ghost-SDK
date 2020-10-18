package state_transition

import (
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/shared/trieutil"
	"github.com/ulule/deepcopier"
	"log"
)

func toByte(str string) []byte {
	ret, _ := hex.DecodeString(str)
	return ret
}

func defaultEth1Data() (*core.ETH1Data, error) {
	trie, err := trieutil.NewTrie(int(params.ChainConfig.DepositContractTreeDepth))
	if err != nil {
		return nil, err
	}
	depositRoot := trie.Root()

	return &core.ETH1Data{
		DepositRoot:          depositRoot[:],
		DepositCount:         0,
		BlockHash:            toByte("8049cb6db44dc94a52a9702779c0b4d5d77164bc56a913da1203563de5193405"),
	}, nil
}

func initBlockHeader() (*core.BlockHeader, error) {
	root, err := ssz.HashTreeRoot(&core.BlockBody{})
	if err != nil {
		return nil, err
	}
	return &core.BlockHeader{
		Slot:                 0,
		ProposerIndex:        0,
		ParentRoot:           nil,
		StateRoot:            nil,
		BodyRoot:             root[:],
	}, nil
}

type StateTestContext struct {
	State *core.State
}

func NewStateTestContext(config *core.ChainConfig, eth1Data *core.ETH1Data, genesisTime uint64) (*StateTestContext, error) {
	var err error
	if eth1Data == nil {
		eth1Data, err = defaultEth1Data()
		if err != nil {
			return nil, err
		}
	}

	initBlockHeader, err := initBlockHeader()
	if err != nil {
		return nil, err
	}

	randaoMixes := make([][]byte, params.ChainConfig.EpochsPerHistoricalVector)
	for i := range randaoMixes {
		randaoMixes[i] = eth1Data.BlockHash
	}

	blockRoots := make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot)
	for i := range blockRoots {
		blockRoots[i] = params.ChainConfig.ZeroHash
	}

	stateRoots := make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot)
	for i := range stateRoots {
		stateRoots[i] = params.ChainConfig.ZeroHash
	}

	genesisValidatorRoot, err := ssz.HashTreeRoot([]*core.Validator{})
	if err != nil {
		return nil, err
	}

	return &StateTestContext{
		State: &core.State{
			GenesisTime:                 genesisTime,
			CurrentSlot:                 0,
			LatestBlockHeader:           initBlockHeader,
			Fork:                        &core.Fork{
				PreviousVersion:      config.GenesisForkVersion,
				CurrentVersion:       config.GenesisForkVersion,
				Epoch:                config.GenesisEpoch,
			},
			BlockRoots:                blockRoots,
			StateRoots:                stateRoots,
			RandaoMix:                 randaoMixes,
			HistoricalRoots:           [][]byte{},
			GenesisValidatorsRoot:     genesisValidatorRoot[:],
			PreviousEpochAttestations: []*core.PendingAttestation{},
			CurrentEpochAttestations:  []*core.PendingAttestation{},
			JustificationBits:         []byte{0},
			PreviousJustifiedCheckpoint: &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			CurrentJustifiedCheckpoint:  &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			FinalizedCheckpoint:         &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			Eth1Data:                    eth1Data,
			Eth1DataVotes:               []*core.ETH1Data{},
			Eth1DepositIndex:            0,
			Validators:                  []*core.Validator{},
			Slashings:                   []uint64{},
		},
	}, nil
}

func (c *StateTestContext) PopulateGenesisValidator(validatorIndexEnd uint64) *StateTestContext {
	if err := bls.Init(bls.BLS12_381); err != nil {
		log.Fatal(err)
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		log.Fatal(err)
	}

	var leaves [][]byte
	deposits := make([]*core.Deposit, validatorIndexEnd)
	for i := uint64(0) ; i < validatorIndexEnd ; i++ {
		sk := &bls.SecretKey{}
		sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", uint64(i)))))

		depositMessage := &core.DepositMessage{
			PublicKey:             sk.GetPublicKey().Serialize(),
			WithdrawalCredentials: []byte("test_withdrawal_cred"),
			Amount:                32 * 1e9, // gwei
		}
		root, err := ssz.HashTreeRoot(depositMessage)
		if err != nil {
			log.Fatal(err)
		}

		// inser into the trie
		depositData := &core.Deposit_DepositData{
			PublicKey:             depositMessage.PublicKey,
			WithdrawalCredentials: depositMessage.WithdrawalCredentials,
			Amount:                depositMessage.Amount,
			Signature:             sk.SignByte(root[:]).Serialize(),
		}
		deposits[i] = &core.Deposit{
			Proof:                nil,
			Data:                 depositData,
		}
		root, err = ssz.HashTreeRoot(depositData)
		if err != nil {
			log.Fatal(err)
		}
		leaves = append(leaves, root[:])
	}

	var trie *trieutil.SparseMerkleTrie
	var err error
	if len(leaves) > 0 {
		trie, err = trieutil.GenerateTrieFromItems(leaves, int(params.ChainConfig.DepositContractTreeDepth))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		trie, err = trieutil.NewTrie(int(params.ChainConfig.DepositContractTreeDepth))
		if err != nil {
			log.Fatal(err)
		}
	}

	depositRoot := trie.Root()
	c.State.Eth1Data.DepositRoot = depositRoot[:]

	// process
	for _, deposit := range deposits {
		// Add validator and balance entries
		v := GetBPFromDeposit(c.State, deposit)
		v.ActivationEpoch = 0
		c.State.Validators = append(c.State.Validators, v)
	}

	// update genesis root
	genesisValidatorRoot, err := ssz.HashTreeRoot(c.State.Validators)
	if err != nil {
		log.Fatal(err)
	}
	c.State.GenesisValidatorsRoot = genesisValidatorRoot[:]

	return c
}

// will generate and save blocks from slot 0 until maxBlocks
func (c *StateTestContext) ProgressSlotsAndEpochs(maxBlocks int) {
	var previousBlockHeader *core.BlockHeader
	previousBlockHeader, err := initBlockHeader()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0 ; i < maxBlocks ; i++ {
		fmt.Printf("progressing block %d\n", i)
		// get proposer
		stateCopy := shared.CopyState(c.State)
		// we increment it before we gt proposer as this is what happens in real block processing
		// in slot > 0 the process slots is called before process block which increments slot by 1
		if i != 0 {
			stateCopy.CurrentSlot ++
		}
		pID, err := shared.GetBlockProposerIndex(stateCopy)
		if err != nil {
			log.Fatal(err)
		}
		sk := []byte(fmt.Sprintf("%d", pID))

		// randao
		randaoReveal, err := signRandao(c.State, sk)
		if err != nil {
			log.Fatal(err)
		}

		// parent
		// replicates what the next process slot does
		if i != 0 {
			stateRoot,err := ssz.HashTreeRoot(c.State)
			if err != nil {
				log.Fatal(err)
			}
			previousBlockHeader.StateRoot =  stateRoot[:]
		}
		parentRoot,err := ssz.HashTreeRoot(previousBlockHeader)
		if err != nil {
			log.Fatal(err)
		}

		block := &core.Block{
			Slot:                 uint64(i),
			Proposer:             pID,
			ParentRoot:           parentRoot[:],
			StateRoot:            params.ChainConfig.ZeroHash,
			Body:                 &core.BlockBody{
				RandaoReveal:         randaoReveal.Serialize(),
				Attestations:         []*core.Attestation{},
			},
		}

		// process
		st := NewStateTransition()

		// compute state root
		root, err := st.ComputeStateRoot(c.State, &core.SignedBlock{
			Block:                block,
			Signature:            []byte{},
		})
		if err != nil {
			log.Fatal(err)
		}
		block.StateRoot = root[:]

		// sign
		blockDomain, err := shared.GetDomain(c.State, params.ChainConfig.DomainBeaconProposer, shared.GetCurrentEpoch(c.State))
		if err != nil {
			log.Fatal(err)
		}
		sig, err := shared.SignBlock(block, sk, blockDomain) // TODO - dynamic domain
		if err != nil {
			log.Fatal(err)
		}

		// execute
		c.State, err = st.ExecuteStateTransition(c.State, &core.SignedBlock{
			Block:                block,
			Signature:            sig.Serialize(),
		})
		if err != nil {
			log.Fatal(err)
		}

		// copy to previousBlockRoot
		previousBlockHeader = &core.BlockHeader{}
		deepcopier.Copy(c.State.LatestBlockHeader).To(previousBlockHeader)
	}
}

func generateAttestations(
	state *core.State,
	howManyBpSig uint64,
	slot uint64,
	sourceCheckpoint *core.Checkpoint,
	targetCheckpoint *core.Checkpoint,
	committeeIdx uint64,
	finalized bool,
	dutyType int32, // 0 - attestation, 1 - proposal, 2 - aggregation
	) []*core.Attestation {

	data := &core.AttestationData{
		Slot:                 slot,
		CommitteeIndex:       committeeIdx,
		BeaconBlockRoot:      []byte("block root"),
		Source:               sourceCheckpoint,
		Target:               targetCheckpoint,
	}

	// sign
	root, err := ssz.HashTreeRoot(data)
	if err != nil {
		return nil
	}

	expectedCommittee, err := shared.GetAttestationCommittee(state, data.Slot, uint64(data.CommitteeIndex))
	if err != nil {
		return nil
	}

	var aggregatedSig *bls.Sign
	aggBits := make(bitfield.Bitlist, len(expectedCommittee)) // for bytes
	signed := uint64(0)
	for i, bpId := range expectedCommittee {
		bp := shared.GetValidator(state, bpId)
		sk := &bls.SecretKey{}
		sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", bp.Id))))

		// sign
		if aggregatedSig == nil {
			aggregatedSig = sk.SignByte(root[:])
		} else {
			aggregatedSig.Add(sk.SignByte(root[:]))
		}
		aggBits.SetBitAt(uint64(i), true)
		signed ++

		if signed >= howManyBpSig {
			break
		}
	}

	return []*core.Attestation{
		{
			Data:            data,
			Signature:       aggregatedSig.Serialize(),
			AggregationBits: aggBits,
		},
	}
}

//func generateTestState(t *testing.T, headSlot int) *core.State {
//	require.NoError(t, bls.Init(bls.BLS12_381))
//	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))
//
//	// block producers
//	bps := make([]*core.Validator, 124)
//	for i := 0 ; i < len(bps) ; i++ {
//		sk := &bls.SecretKey{}
//		sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", uint64(i)))))
//
//		bps[i] = &core.Validator {
//			Id:      uint64(i),
//			EffectiveBalance:   32,
//			Balance : 32,
//			Slashed: false,
//			Active:  true,
//			PubKey:  sk.GetPublicKey().Serialize(),
//		}
//	}
//
//	// vaults (pool)
//	//for i := 0 ; i < len(pools) ; i++ {
//	//	executors := make([]uint64, params.ChainConfig.VaultSize)
//	//	for j := 0 ; j < int(params.ChainConfig.VaultSize) ; j++ {
//	//		executors[j] = bps[i*int(params.ChainConfig.VaultSize) + j].GetId()
//	//	} // no need to sort as they are already
//	//
//	//	sk := &bls.SecretKey{}
//	//	sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", uint64(i)))))
//	//
//	//	pools[i] = &core.Pool{
//	//		Id:              uint64(i),
//	//		SortedCommittee: executors,
//	//		PubKey:          sk.GetPublicKey().Serialize(),
//	//		Active:true,
//	//	}
//	//}
//	ret := &core.State {
//		CurrentSlot: 0,
//		Validators:  bps,
//		RandaoMix:          	   make([][]byte, params.ChainConfig.EpochsPerHistoricalVector),
//		BlockRoots:                make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot),
//		StateRoots:                make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot),
//		PreviousEpochAttestations: []*core.PendingAttestation{},
//		CurrentEpochAttestations:  []*core.PendingAttestation{},
//		JustificationBits:         []byte{0},
//		PreviousJustifiedCheckpoint: &core.Checkpoint{
//			Epoch:                0,
//			Root:                 params.ChainConfig.ZeroHash, // TODO is it zero hash?
//		},
//		CurrentJustifiedCheckpoint: &core.Checkpoint{
//			Epoch:                0,
//			Root:                 params.ChainConfig.ZeroHash, // TODO is it zero hash?
//		},
//	}
//
//	ret, err := generateAndApplyBlocks(ret, headSlot)
//	if err != nil {
//		fmt.Printf("%s\n", err.Error())
//		return nil
//	}
//
//	return ret
//}

// will populate the prev and current pending attestations for processing
func populateJustificationAndFinalization(
		state *core.State,
		epoch uint64,
		endSlot uint64, // end slot for attestations
		participationRate float64,
		targetCheckpoint *core.Checkpoint,
	) error {

	slotPointer := epoch * params.ChainConfig.SlotsInEpoch // points at first slot
	pendingAttArray := make([]*core.PendingAttestation, 0)
	setPendingAttArray := func(att []*core.PendingAttestation) {
		if epoch == shared.GetCurrentEpoch(state) {
			state.CurrentEpochAttestations = att
		} else {
			state.PreviousEpochAttestations = att
		}
	}

	for slotPointer <= endSlot {

		for cIndx := uint64(0) ; cIndx < shared.GetCommitteeCountPerSlot(state, slotPointer) ; cIndx ++ {
			committee, err := shared.GetAttestationCommittee(state, slotPointer, cIndx)
			if err != nil {
				return err
			}
			targetParticipation := uint64(float64(len(committee)) * participationRate)

			aggBits := make(bitfield.Bitlist, params.ChainConfig.MaxValidatorsPerCommittee)
			for i := uint64(0); i < targetParticipation; i ++ {
				aggBits.SetBitAt(i, true)
			}

			pendingAttArray = append(pendingAttArray, &core.PendingAttestation{
				AggregationBits:      aggBits,
				Data:                 &core.AttestationData{
					Slot:                 slotPointer,
					CommitteeIndex:       cIndx,
					Target:               targetCheckpoint,
				},
			})
		}
		slotPointer ++
	}

	setPendingAttArray(pendingAttArray)

	return nil
}

func signRandao(state *core.State, sk []byte) (*bls.Sign, error) {
	// We bump the slot by one to accurately calculate the epoch as when this
	// randao reveal will be verified the state.CurrentSlot will be +1
	copyState := shared.CopyState(state)
	copyState.CurrentSlot ++

	data, domain, err := RANDAOSigningData(copyState)
	if err != nil {
		return nil, err
	}
	return shared.SignRandao(data, domain, sk)
}