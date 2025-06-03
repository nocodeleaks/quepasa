// Package library provides utility functions for generating checksums.
package library

import (
	"crypto/md5" // For MD5 hashing
	"fmt"        // For string formatting
	"hash/crc32" // For CRC32 checksums
)

// GenerateCRC32ChecksumString calculates the CRC32 (IEEE) checksum for a []byte
// and returns it as an 8-character, zero-padded hexadecimal string.
//
// Parameters:
//
//	data: The byte slice to calculate the checksum for.
//
// Returns:
//
//	A string representing the checksum in hexadecimal format (e.g., "0000a1b2").
func GenerateCRC32ChecksumString(data []byte) string {
	// Calculate the IEEE CRC32 checksum.
	// crc32.IEEE is a common polynomial table ensuring consistent results.
	checksumValue := crc32.ChecksumIEEE(data)

	// Format the uint32 checksumValue as an 8-character hexadecimal string,
	// padded with leading zeros if necessary.
	return fmt.Sprintf("%08x", checksumValue)
}

// GenerateMD5ChecksumString calculates the MD5 checksum for a []byte
// and returns it as a 32-character hexadecimal string.
//
// Parameters:
//
//	data: The byte slice to calculate the checksum for.
//
// Returns:
//
//	A string representing the MD5 checksum in hexadecimal format (e.g., "d41d8cd98f00b204e9800998ecf8427e").
func GenerateMD5ChecksumString(data []byte) string {
	// Create a new MD5 hasher instance.
	hasher := md5.New()

	// Write the input data to the hasher.
	// The hasher.Write method never returns an error.
	hasher.Write(data)

	// Calculate the MD5 sum. Sum appends the current hash to b and returns the resulting slice.
	// Passing nil causes Sum to allocate a new slice.
	checksumBytes := hasher.Sum(nil)

	// Format the byte slice checksum as a hexadecimal string.
	return fmt.Sprintf("%x", checksumBytes)
}
