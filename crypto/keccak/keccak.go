/*

Copyright 2012-2013 The CryptoNote Developers
Copyright 2014-2018 The Monero Developers
Copyright 2018 The TurtleCoin Developers

Please see the included LICENSE file for more information

*/

package keccak

import (
	"encoding/binary"
)

func keccakf(state []uint64, rounds int) {
	var t uint64
	var bc [5]uint64

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

// Keccak computes the keccak hash of given byte length
func Keccak(input []byte, outputSize int) []byte {

	state := make([]uint64, 25)

	var rsiz int

	if len(state) * 8 == outputSize {
		rsiz = hashDataArea
	} else {
		rsiz = 200 - 2*outputSize
	}

	inputLength := len(input)

	for inputLength >= rsiz {
		for i := 0; i < rsiz; i += 8 {
			state[i/8] ^= binary.LittleEndian.Uint64(input[i:(i + 8)])
		}
		keccakf(state, keccakRounds)
		inputLength -= rsiz
	}

	temp := make([]byte, 144)

	for i := 0; i < inputLength; i++ {
		temp[i] = input[i]
	}

	temp[inputLength] = 1
	inputLength++

	for i := inputLength; i < rsiz; i++ {
		temp[i] = 0
	}

	temp[rsiz-1] |= 0x80
	temp[rsiz] = 1

	for i := 0; i < rsiz; i += 8 {
		state[i/8] ^= binary.LittleEndian.Uint64(temp[i:(i + 8)])
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

func rotl64(x uint64, y uint64) uint64 {
	return x<<y | x>>(64-y)
}
