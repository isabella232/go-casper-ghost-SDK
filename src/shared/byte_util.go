package shared

import "encoding/binary"

func SliceToByte32(slice []byte) [32]byte {
	var arr [32]byte
	copy(arr[:], slice[:32])
	return arr
}

// Bytes4 returns integer x to bytes in little-endian format, x.to_bytes(4, 'little').
func Bytes4(x uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, x)
	return bytes[:4]
}

// Bytes8 returns integer x to bytes in little-endian format, x.to_bytes(8, 'little').
func Bytes8(x uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, x)
	return bytes
}

// ToBytes4 is a convenience method for converting a byte slice to a fix
// sized 4 byte array. This method will truncate the input if it is larger
// than 4 bytes.
func ToBytes4(x []byte) [4]byte {
	var y [4]byte
	copy(y[:], x)
	return y
}