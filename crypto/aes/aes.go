/*
 * ---------------------------------------------------------------------------
 * OpenAES License
 * ---------------------------------------------------------------------------
 * Copyright (c) 2012, Nabil S. Al Ramli, www.nalramli.com
 * Copyright (c) 2018, The TurtleCoin Developers
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *   - Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *   - Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 * ---------------------------------------------------------------------------
 */

package aes

// Offset = offset of input array to access
func PseudoEncryptECB(keys, input []byte, offset int) {
	for i := 0; i < 10; i++ {
		EncryptionRound(keys, input, offset, blockSize*i)
	}
}

func EncryptionRound(keys, data []byte, offset, keyOffset int) {
	for i := 0; i < blockSize; i++ {
		data[i+offset] = subByte(data[i+offset])
	}

	shiftRows(data, offset)

	mixColumns(data, offset)

	for i := 0; i < blockSize; i++ {
		// Select the appropriate key to use via the offset
		data[i+offset] ^= keys[i+offset]
	}
}

func ExpandKey(key []byte) []byte {
	keyBase := len(key) / roundKeyLength
	numKeys := keyBase + roundBase

	expandedKeyLength := numKeys * roundKeyLength * columnLength

	expanded := make([]byte, expandedKeyLength)

	// First key is a direct copy of input key
	copy(expanded, key)

	// Apply expand key algorithm for remaining keys
	for i := keyBase; i < numKeys*roundKeyLength; i++ {
		temp := make([]byte, columnLength)

		/* Copy column length bytes from expanded to tmp, with an
		   offset of (i - 1) * RoundKeyLength from expanded. */
		for j := 0; j < columnLength; j++ {
			temp[j] = expanded[j+(i-1)*roundKeyLength]
		}

		if i%keyBase == 0 {
			temp = rotLeft(temp)

			for j := 0; j < columnLength; j++ {
				temp[j] = subByte(temp[j])
			}

			temp[0] = byte((temp[0] ^ gf8[(i/keyBase)-1]))
		} else if keyBase > 6 && (i%keyBase) == 4 {
			for j := 0; j < columnLength; j++ {
				temp[j] = subByte(temp[j])
			}
		}

		for j := 0; j < columnLength; j++ {
			index := ((i - keyBase) * roundKeyLength) + j

			expanded[(i*roundKeyLength)+j] = byte((expanded[index] ^ temp[j]))
		}
	}

	return expanded
}

func shiftRows(input []byte, offset int) {
	temp := make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		index := (i * 5) % blockSize
		temp[i] = input[offset+index]
	}

	// Copy temp array to output
	for i := 0; i < len(temp); i++ {
		input[i+offset] = temp[i]
	}
}

func mixColumns(input []byte, offset int) {
	temp := make([]byte, blockSize)

	for i := 0; i < blockSize; i += columnLength {
		temp[i] = byte(gfMul(input[i+offset], 2) ^ gfMul(input[i+1+offset], 3) ^ input[i+2+offset] ^ input[i+3+offset])
		temp[i+1] = byte(input[i+offset] ^ gfMul(input[i+1+offset], 2) ^ gfMul(input[i+2+offset], 3) ^ input[i+3+offset])
		temp[i+2] = byte(input[i+offset] ^ input[i+1+offset] ^ gfMul(input[i+2+offset], 2) ^ gfMul(input[i+3+offset], 3))
		temp[i+3] = byte(gfMul(input[i+offset], 3) ^ input[i+1+offset] ^ input[i+2+offset] ^ gfMul(input[i+3+offset], 2))
	}

	for i := 0; i < len(temp); i++ {
		input[i+offset] = temp[i]
	}
}

func gfMul(left, right byte) byte {
	x := left
	y := left

	x &= 0x0f
	y &= 0xf0

	y >>= 4

	switch right {
	case 0x02:
		return gfMul2[y][x]
	case 0x03:
		return gfMul3[y][x]
	case 0x09:
		return gfMul9[y][x]
	case 0x0b:
		return gfMulb[y][x]
	case 0x0d:
		return gfMuld[y][x]
	case 0x0e:
		return gfMule[y][x]
	default:
		return left
	}
}

func rotLeft(input []byte) []byte {
	output := make([]byte, columnLength)

	for i := 0; i < columnLength; i++ {
		output[i] = input[(i+1)%columnLength]
	}

	return output
}

func subByte(input byte) byte {
	x := input
	y := input

	x &= 0x0f
	y &= 0xf0

	y >>= 4

	return subByteValue[y][x]
}
