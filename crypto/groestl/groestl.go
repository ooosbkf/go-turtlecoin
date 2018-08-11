/*
Copyright 2014 Diego Alejandro Gómez <diego.gomezy@udea.edu.co>
Copyright 2018 The TurtleCoin Developers
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.
*/

package groestl

import (
	"encoding/binary"
	"unsafe"
)

func doubling(x byte) byte {
	if x&0x80 == 0x80 {
		return byte((x << 1) ^ 0x1b)
	}

	return byte(x << 1)
}

func xor(op1, op2 []byte) []byte {
	result := make([]byte, len(op1))
	for i := 0; i < len(result); i++ {
		result[i] = byte(op1[i] ^ op2[i])
	}

	return result
}

func shiftRow(row []byte, shift int) []byte {
	newRow := make([]byte, len(row))

	for i := 0; i < len(row)-shift; i++ {
		newRow[i] = row[i+shift]
	}

	for i := 0; i < shift; i++ {
		newRow[i+len(row)-shift] = row[i]
	}

	return newRow
}

func truncation(x []byte) []byte {
	nBytes := 32

	result := make([]byte, nBytes)
	for i := 0; i < nBytes; i++ {
		result[i] = x[i+len(x)-nBytes]
	}

	return result
}

func matrixToBytesMap(input [][]byte) []byte {
	nBytes := 64
	k := 0

	result := make([]byte, nBytes)
	for j := 0; j < nBytes/8; j++ {
		for i := 0; i < 8; i++ {
			result[k] = input[i][j]
			k++
		}
	}

	return result
}

func bytestoMatrixMap(input []byte) [][]byte {
	nColumns := 8
	k := 0

	result := [][]byte{
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns),
		make([]byte, nColumns)}

	for j := 0; j < nColumns; j++ {
		for i := 0; i < 8; i++ {
			result[i][j] = input[k]
			k++
		}
	}

	return result
}

func mixBytes(state [][]byte) {
	nColumns := 8

	x := make([]byte, 8)
	y := make([]byte, 8)
	z := make([]byte, 8)

	for j := 0; j < nColumns; j++ {
		x[0] = byte(state[0][j] ^ state[(0+1)%8][j])
		x[1] = byte(state[1][j] ^ state[(1+1)%8][j])
		x[2] = byte(state[2][j] ^ state[(2+1)%8][j])
		x[3] = byte(state[3][j] ^ state[(3+1)%8][j])
		x[4] = byte(state[4][j] ^ state[(4+1)%8][j])
		x[5] = byte(state[5][j] ^ state[(5+1)%8][j])
		x[6] = byte(state[6][j] ^ state[(6+1)%8][j])
		x[7] = byte(state[7][j] ^ state[(7+1)%8][j])
		y[0] = byte(x[0] ^ x[(0+3)%8])
		y[1] = byte(x[1] ^ x[(1+3)%8])
		y[2] = byte(x[2] ^ x[(2+3)%8])
		y[3] = byte(x[3] ^ x[(3+3)%8])
		y[4] = byte(x[4] ^ x[(4+3)%8])
		y[5] = byte(x[5] ^ x[(5+3)%8])
		y[6] = byte(x[6] ^ x[(6+3)%8])
		y[7] = byte(x[7] ^ x[(7+3)%8])
		z[0] = byte(x[0] ^ x[(0+2)%8] ^ state[(0+6)%8][j])
		z[1] = byte(x[1] ^ x[(1+2)%8] ^ state[(1+6)%8][j])
		z[2] = byte(x[2] ^ x[(2+2)%8] ^ state[(2+6)%8][j])
		z[3] = byte(x[3] ^ x[(3+2)%8] ^ state[(3+6)%8][j])
		z[4] = byte(x[4] ^ x[(4+2)%8] ^ state[(4+6)%8][j])
		z[5] = byte(x[5] ^ x[(5+2)%8] ^ state[(5+6)%8][j])
		z[6] = byte(x[6] ^ x[(6+2)%8] ^ state[(6+6)%8][j])
		z[7] = byte(x[7] ^ x[(7+2)%8] ^ state[(7+6)%8][j])
		state[0][j] = byte(doubling(byte(doubling(y[(0+3)%8])^z[(0+7)%8])) ^ z[(0+4)%8])
		state[1][j] = byte(doubling(byte(doubling(y[(1+3)%8])^z[(1+7)%8])) ^ z[(1+4)%8])
		state[2][j] = byte(doubling(byte(doubling(y[(2+3)%8])^z[(2+7)%8])) ^ z[(2+4)%8])
		state[3][j] = byte(doubling(byte(doubling(y[(3+3)%8])^z[(3+7)%8])) ^ z[(3+4)%8])
		state[4][j] = byte(doubling(byte(doubling(y[(4+3)%8])^z[(4+7)%8])) ^ z[(4+4)%8])
		state[5][j] = byte(doubling(byte(doubling(y[(5+3)%8])^z[(5+7)%8])) ^ z[(5+4)%8])
		state[6][j] = byte(doubling(byte(doubling(y[(6+3)%8])^z[(6+7)%8])) ^ z[(6+4)%8])
		state[7][j] = byte(doubling(byte(doubling(y[(7+3)%8])^z[(7+7)%8])) ^ z[(7+4)%8])
	}
}

func shiftBytes(state [][]byte, permutation byte) {
	var sigma []byte

	if permutation == 0 {
		sigma = []byte{
			0, 1, 2, 3, 4, 5, 6, 7}
	} else {
		sigma = []byte{
			1, 3, 5, 7, 0, 2, 4, 6}
	}

	for i := 0; i < 8; i++ {
		state[i] = shiftRow(state[i], int(sigma[i]))
	}
}

func subBytes(state [][]byte) {
	nColumns := 8

	for i := 0; i < 8; i++ {
		for j := 0; j < nColumns; j++ {
			state[i][j] = sBox[state[i][j]]
		}
	}
}

func addRoundConstant(state [][]byte, r byte, permutation byte) {
	nColumns := 8

	if permutation == 0 {
		c := []byte{
			(byte)(0x00 ^ r), (byte)(0x10 ^ r), (byte)(0x20 ^ r), (byte)(0x30 ^ r),
			(byte)(0x40 ^ r), (byte)(0x50 ^ r), (byte)(0x60 ^ r), (byte)(0x70 ^ r),
			(byte)(0x80 ^ r), (byte)(0x90 ^ r), (byte)(0xa0 ^ r), (byte)(0xb0 ^ r),
			(byte)(0xc0 ^ r), (byte)(0xd0 ^ r), (byte)(0xe0 ^ r), (byte)(0xf0 ^ r)}

		for j := 0; j < nColumns; j++ {
			state[0][j] = byte(state[0][j] ^ c[j])
		}
	} else {
		c := []byte{
			(byte)(0xff ^ r), (byte)(0xef ^ r), (byte)(0xdf ^ r), (byte)(0xcf ^ r),
			(byte)(0xbf ^ r), (byte)(0xaf ^ r), (byte)(0x9f ^ r), (byte)(0x8f ^ r),
			(byte)(0x7f ^ r), (byte)(0x6f ^ r), (byte)(0x5f ^ r), (byte)(0x4f ^ r),
			(byte)(0x3f ^ r), (byte)(0x2f ^ r), (byte)(0x1f ^ r), (byte)(0x0f ^ r)}

		for i := 0; i < 7; i++ {
			for j := 0; j < nColumns; j++ {
				state[i][j] = byte(state[i][j] ^ 0xff)
			}
		}
		for j := 0; j < nColumns; j++ {
			state[7][j] = byte(state[7][j] ^ c[j])
		}
	}
}

func p(input []byte) []byte {
	R := 10
	state := bytestoMatrixMap(input)

	for r := byte(0); r < byte(R); r++ {
		addRoundConstant(state, r, byte(0))
		subBytes(state)
		shiftBytes(state, 0)
		mixBytes(state)
	}

	return matrixToBytesMap(state)
}

func q(input []byte) []byte {
	R := 10
	state := bytestoMatrixMap(input)

	for r := byte(0); r < byte(R); r++ {
		addRoundConstant(state, r, byte(1))
		subBytes(state)
		shiftBytes(state, 1)
		mixBytes(state)
	}

	return matrixToBytesMap(state)
}

func compression(h, m []byte) []byte {
	return xor(xor(p(xor(h, m)), q(m)), h)
}

const intSize int = int(unsafe.Sizeof(0))

func getEndian() (ret bool) {
	var i = 0x1
	bs := (*[intSize]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		return true
	}

	return false
}

func reverse(input []byte) []byte {
	temp := make([]byte, len(input))
	j := 0
	for i := len(input) - 1; i >= 0; i-- {
		temp[j] = input[i]
		j++
	}
	return temp
}

// Hash calculates the Groestl hash
// for given input
func Hash(input []byte) []byte {
	/*
	 The following variables follow the same naming convention used in the
	 specification:
	     l: length of each message block, in bits.
	     t: number of blocks the whole message uses.
	     T: total number of blocks.
	     N: number of bits in the input message.
	     w: number of zeroes for padding.
	*/
	l := bufferSize * 8
	N := uint64(0)
	t := uint64(0)
	T := uint64(0)
	w := uint64(0)

	h := make([]byte, bufferSize)
	for i := 0; i < bufferSize; i++ {
		h[i] = iv[i]
	}

	buffer := make([]byte, bufferSize)

	bytesRead := 0

	for i := 0; i < len(input); i += bufferSize {
		if len(input)-i < bufferSize {
			bytesRead = len(input) - i
			for j := 0; j < len(input)-i; j++ {
				buffer[j] = input[j+i]
			}
			break
		}

		for j := 0; j < bufferSize; j++ {
			buffer[j] = input[j+i]
		}

		N = N + uint64(bufferSize*8)
		h = compression(h, buffer)
		t++
	}

	N = N + uint64(bytesRead*8)
	w = uint64((((-int64(N) - int64(65)) % int64(l)) + int64(l)) % int64(l))
	T = (N + w + 65) / uint64(l)

	blocksLeft := int(T - t)
	buffer[bytesRead] = 0x80
	noOfZeroBytes := bufferSize - bytesRead - 1
	zeroes := make([]byte, noOfZeroBytes)

	for i := 0; i < noOfZeroBytes; i++ {
		buffer[i+bytesRead+1] = zeroes[i]
	}

	for i := uint(0); i < uint(blocksLeft); i++ {
		if i == uint(blocksLeft-1) {
			bytes := make([]byte, 8)
			binary.BigEndian.PutUint64(bytes, T)
			if !getEndian() {
				bytes = reverse(bytes)
			}
			for j := 0; j < 8; j++ {
				buffer[j+bufferSize-8] = bytes[j]
			}
		}
		h = compression(h, buffer)
		buffer = make([]byte, bufferSize)
	}

	h = truncation(xor(p(h), h))

	return h
}
