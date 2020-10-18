package core

import "bytes"

func CheckpointsEqual(l *Checkpoint, r *Checkpoint) bool {
	return l.Epoch == r.Epoch && bytes.Equal(l.Root, r.Root)
}

// returns true if equal
func AttestationDataEqual(att1 *AttestationData, att2 *AttestationData) bool {
	return att1.Slot == att2.Slot &&
		CheckpointsEqual(att1.Target, att2.Target) &&
		CheckpointsEqual(att1.Source, att2.Source) &&
		bytes.Equal(att1.BeaconBlockRoot, att2.BeaconBlockRoot)
}

func BlockHeaderEqual(head1 *PoolBlockHeader, head2 *PoolBlockHeader) bool {
	return head1.ProposerIndex == head2.ProposerIndex &&
		head1.Slot == head2.Slot &&
		bytes.Equal(head1.StateRoot, head2.StateRoot) &&
		bytes.Equal(head1.ParentRoot, head2.ParentRoot) &&
		bytes.Equal(head1.BodyRoot, head2.BodyRoot)
}