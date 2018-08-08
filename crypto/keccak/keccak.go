/*

Copyright 2011 Markku-Juhani O. Saarinen
Copyright 2012-2013 The CryptoNote Developers
Copyright 2014-2018 The Monero Developers
Copyright 2018 The TurtleCoin Developers

Please see the included LICENSE file for more information

*/

// This is pre NIST keccak before the sha-3 revisions

package keccak

import (
	"encoding/binary"
	"fmt"
	"math"
)

func keccakf(state []uint64, rounds int) {
	var t uint64
	var bc [5]uint64

	if rounds == -1 {
		rounds = keccakRounds
	}

	for round := 0; round < rounds; round++ {

		// Theta
		for i := 0; i < 5; i++ {
			bc[i] = state[i] ^ state[i+5] ^ state[i+10] ^ state[i+15] ^ state[i+20]
		}

		for i := 0; i < 5; i++ {
			t = bc[(i+4)%5] ^ rotl64(bc[(i+1)%5], 1)

			for j := 0; j < 25; j += 5 {
				state[i+j] ^= t
			}
		}

		// Rho Pi
		t = state[1]

		for i := 0; i < 24; i++ {
			j := keccakfPiln[i]
			bc[0] = state[j]
			state[j] = rotl64(t, uint64(keccakfRotc[i]))
			t = bc[0]
		}

		//Chi
		for j := 0; j < 25; j += 5 {
			for i := 0; i < 5; i++ {
				bc[i] = state[i+j]
			}

			for i := 0; i < 5; i++ {
				state[i+j] ^= (^bc[(i+1)%5]) & bc[(i+2)%5]
			}
		}

		//Iota
		state[0] ^= keccakfRc[round]
	}
}

// Compute a hash of length outputSize from input
func keccak(input []byte, outputSize int) []byte {

	state := make([]uint64, 25)

	rsiz := hashDataArea

	if outputSize != 200 {
		rsiz = 200 - 2*outputSize
	}

	rsizw := rsiz / 8

	inputLength := len(input)

	for inputLength >= rsiz {
		for i := 0; i < rsizw; i++ {
			/* Read 8 bytes as a ulong, need to multiply i by
			8 because we're reading chunks of 8 at once */
			state[i] ^= binary.LittleEndian.Uint64(input[i*8 : (i*8 + 8)])
		}
		keccakf(state, -1)
		inputLength -= rsiz
	}

	temp := make([]byte, 144)

	/* Copy inputLength bytes from input to tmp at an offset of
	   offset from input */
	for i := 0; i < inputLength; i++ {
		temp[i] = input[i]
	}

	temp[inputLength] = 1
	inputLength++

	/* Zero (rsiz - inputLength) bytes in tmp, at an offset of
	   inputLength */
	for i := inputLength; i < rsiz; i++ {
		temp[i] = 0
	}

	temp[rsiz-1] |= 0x80
	temp[rsiz] = 1

	for i := 0; i < rsizw; i++ {
		/* Read 8 bytes as a ulong - need to read at (i * 8) because
		   we're reading chunks of 8 at once, rather than overlapping
		   chunks of 8 */
		state[i] ^= binary.LittleEndian.Uint64(temp[i*8 : (i*8 + 8)])
	}

	keccakf(state, keccakRounds)

	// following part is similar to memcpy in C/C++
	temp1 := make([]byte, outputSize/4)
	output := []byte{}

	for i := 0; i < outputSize; i += 8 {
		binary.LittleEndian.PutUint64(temp1, state[i/8])
		output = append(output, temp1...)
	}

	return output

}

// Keccak hashes the given input with keccak, into an output hash of 32 bytes.
// Copies outputLength bytes of the output and returns it. Output
// length cannot be larger than 32.
func Keccak(input []byte, outputLength int) []byte {

	if outputLength > 32 {
		fmt.Println("Output length must be 32 bytes or less")
	}

	if outputLength == -1 {
		outputLength = 32
	}

	result := keccak(input, 32)

	output := make([]byte, outputLength)

	// Don't overflow input array
	for i := 0; i < int(math.Min(float64(outputLength), float64(32))); i++ {
		output[i] = result[i]
	}

	return output
}

// Keccak1600 hashes the given input with keccak,
// into an output hash of 200 bytes.
func Keccak1600(input []byte) []byte {

	return keccak(input, 200)

}

func rotl64(x uint64, y uint64) uint64 {
	return x<<y | x>>(64-y)
}
