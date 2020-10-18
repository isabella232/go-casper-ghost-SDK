package shared

import (
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-ssz"
)

func SignBlock(block *core.PoolBlock, sk []byte, domain []byte) (*bls.Sign, error) {
	root, err := BlockSigningRoot(block, domain)
	if err != nil {
		return nil, err
	}

	privKey := bls.SecretKey{}
	err = privKey.SetHexString(hex.EncodeToString(sk))
	if err != nil {
		return nil, err
	}
	sig := privKey.SignByte(root[:])
	return sig, nil
}

func VerifyBlockSigningRoot(block *core.PoolBlock, pubKey []byte, sigByts []byte, domain []byte) error {
	root, err := BlockSigningRoot(block, domain)
	if err != nil {
		return err
	}

	res, err := VerifySignature(root[:], pubKey, sigByts)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("block sig not verified")
	}
	return nil
}

func BlockSigningRoot(block *core.PoolBlock, domain []byte) ([32]byte, error) {
	root, err := ssz.HashTreeRoot(block)
	if err != nil {
		return [32]byte{}, err
	}
	container := struct {
		ObjectRoot []byte
		Domain []byte
	}{
		root[:],
		domain,
	}
	return ssz.HashTreeRoot(container)
}

func RandaoSigningRoot(data []byte, domain []byte) ([32]byte, error) {
	container := struct {
		ObjectRoot []byte
		Domain []byte
	}{
		data,
		domain,
	}
	return ssz.HashTreeRoot(container)
}

func SignRandao(data []byte, domain []byte, sk []byte) (*bls.Sign, error) {
	root, err := RandaoSigningRoot(data, domain)
	if err != nil {
		return nil, err
	}

	privKey := bls.SecretKey{}
	err = privKey.SetHexString(hex.EncodeToString(sk))
	if err != nil {
		return nil, err
	}
	sig := privKey.SignByte(root[:])
	return sig, nil
}

func VerifyRandaoRevealSignature(data []byte, domain []byte, pubKey []byte, sigByts []byte) (bool, error)  {
	root, err := RandaoSigningRoot(data, domain)
	if err != nil {
		return false, err
	}
	return VerifySignature(root[:], pubKey, sigByts)
}

func VerifySignature(root []byte, pubKey []byte, sigByts []byte) (bool, error) {
	return VerifyAggregateSignature(root, [][]byte{pubKey}, sigByts)
}

func VerifyAggregateSignature(root []byte, pubkeys [][]byte, sigByts []byte) (bool, error) {
	// pks
	pks := []bls.PublicKey{}
	for _, pk := range pubkeys {
		_pk := bls.PublicKey{}
		err := _pk.Deserialize(pk)
		if err != nil {
			return false, err
		}
		pks = append(pks, _pk)
	}

	// sig
	sig := &bls.Sign{}
	err := sig.Deserialize(sigByts)
	if err != nil {
		return false, err
	}

	// verify
	if !sig.FastAggregateVerify(pks, root) {
		return false, nil
	}
	return true, nil
}

// Spec pseudocode definition:
//  def get_domain(state: BeaconState, domain_type: DomainType, epoch: Epoch=None) -> GetDomain:
//    """
//    Return the signature domain (fork version concatenated with domain type) of a message.
//    """
//    epoch = get_current_epoch(state) if epoch is None else epoch
//    fork_version = state.fork.previous_version if epoch < state.fork.epoch else state.fork.current_version
//    return compute_domain(domain_type, fork_version, state.genesis_validators_root)
func GetDomain(state *core.State, domainType []byte, epoch uint64) ([]byte, error) {
	epoch = GetCurrentEpoch(state)
	var forkVersion []byte
	if epoch < state.Fork.Epoch {
		forkVersion = state.Fork.PreviousVersion
	} else {
		forkVersion = state.Fork.CurrentVersion
	}

	return ComputeDomain(domainType, forkVersion, state.GenesisValidatorsRoot)
}

// def compute_domain(domain_type: DomainType, fork_version: Version=None, genesis_validators_root: Root=None) -> GetDomain:
//    """
//    Return the domain for the ``domain_type`` and ``fork_version``.
//    """
//    if fork_version is None:
//        fork_version = GENESIS_FORK_VERSION
//    if genesis_validators_root is None:
//        genesis_validators_root = Root()  # all bytes zero by default
//    fork_data_root = compute_fork_data_root(fork_version, genesis_validators_root)
//    return GetDomain(domain_type + fork_data_root[:28])
func ComputeDomain(domainType []byte, forkVersion []byte, genesisValidatorRoot []byte) ([]byte, error) {
	domainBytes := [4]byte{}
	copy(domainBytes[:], domainType[0:4])

	if forkVersion == nil {
		forkVersion = params.ChainConfig.GenesisForkVersion
	}
	if genesisValidatorRoot == nil {
		genesisValidatorRoot = params.ChainConfig.ZeroHash
	}
	forkBytes := make([]byte, 4)
	copy(forkBytes[:], forkVersion)
	forkDataRoot, err := ComputeForkDataRoot(forkVersion, genesisValidatorRoot)
	if err != nil {
		return nil, err
	}

	var b []byte
	b = append(b, domainType[:4]...)
	b = append(b, forkDataRoot[:28]...)
	return b, nil
}

/**
def compute_fork_data_root(current_version: Version, genesis_validators_root: Root) -> Root:
    """
    Return the 32-byte fork data root for the ``current_version`` and ``genesis_validators_root``.
    This is used primarily in signature domains to avoid collisions across forks/chains.
    """
    return hash_tree_root(ForkData(
        current_version=current_version,
        genesis_validators_root=genesis_validators_root,
    ))
 */
func ComputeForkDataRoot(version []byte, root []byte) ([32]byte, error) {
	return ssz.HashTreeRoot(&core.ForkData{
		CurrentVersion:       version,
		GenesisValidatorRoot: root,
	})
}

/**
def compute_fork_digest(current_version: Version, genesis_validators_root: Root) -> ForkDigest:
    """
    Return the 4-byte fork digest for the ``current_version`` and ``genesis_validators_root``.
    This is a digest primarily used for domain separation on the p2p layer.
    4-bytes suffices for practical separation of forks/chains.
    """
    return ForkDigest(compute_fork_data_root(current_version, genesis_validators_root)[:4])
 */
func ComputeForkDigest(version []byte, root []byte) ([4]byte, error) {
	dataRoot, err := ComputeForkDataRoot(version, root)
	if err != nil {
		return [4]byte{}, err
	}
	return ToBytes4(dataRoot[:]), nil
}

/**
def compute_signing_root(ssz_object: SSZObject, domain: Domain) -> Root:
    """
    Return the signing root for the corresponding signing data.
    """
    return hash_tree_root(SigningData(
        object_root=hash_tree_root(ssz_object),
        domain=domain,
    ))
 */
func ComputeSigningRoot(obj interface{}, domain []byte) ([32]byte, error) {
	container := struct {
		ObjectRoot interface{}
		Domain []byte
	}{
		obj,
		domain,
	}
	return ssz.HashTreeRoot(container)
}