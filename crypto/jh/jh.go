/*
Copyright 2011 Hongjun Wu
Copyright 2018 The TurtleCoin Developers

Please see the included LICENSE file for more information.
*/

package jh

type hashState struct {
	H             [128]byte
	A             [256]byte
	roundConstant [64]byte
	buffer        [64]byte
	dataInBuffer  uint64
	dataBitLen    uint64
}

// Hash calculates the JH hash
// of given input
func Hash(input []byte) []byte {
	var state = new(hashState)
	state.dataInBuffer = 0
	state.dataBitLen = 0

	initialize(state)

	update(state, input)

	return final(state)
}

func initialize(state *hashState) {
	state.H[0] = 1
	f8(state)
}

// Hash each 512-bit message block, except the last partial block
func update(state *hashState, input []byte) {
	dataBitLen := uint64(len(input)) * 8

	state.dataBitLen += dataBitLen

	// The starting address for the data to be compressed
	index := uint64(0)

	/* If there is remaining data in the buffer, fill it to a full
	   message block first */

	/* We assume that the size of the data in the buffer is the
	   multiple of 8 bits if it is not at the end of a message */

	/* There is data in the buffer, but the incoming data is
	   insufficient for a full block */
	if (state.dataInBuffer > 0) && ((state.dataInBuffer + dataBitLen) < 512) {
		plus := 0
		if (dataBitLen & 7) != 0 {
			plus = 1
		}

		/* Copy (64 - dataInBuffer >> 3) bytes + plus from input to
		   state.buffer at an offset of dataInBuffer >> 3 in buffer */
		for i := 0; i < (int(64-(state.dataInBuffer>>3)) + plus); i++ {
			state.buffer[i+(int(state.dataInBuffer>>3))] = input[i]
		}

		state.dataInBuffer += dataBitLen
		dataBitLen = 0
	}

	/* There is data in the buffer, and the incoming data is
	   sufficient for a full block */
	if (state.dataInBuffer > 0) && ((state.dataInBuffer + dataBitLen) >= 512) {
		/* Copy (64 - dataInBuffer >> 3) bytes from input to
		   state.buffer at an offset of dataInBuffer >> 3 in buffer */
		for i := 0; i < int((64 - (state.dataInBuffer >> 3))); i++ {
			state.buffer[i+int((state.dataInBuffer>>3))] = input[i]
		}

		index = 64 - (state.dataInBuffer >> 3)

		dataBitLen -= (512 - state.dataInBuffer)

		f8(state)

		state.dataInBuffer = 0
	}

	// Hash the remaining full message blocks
	for dataBitLen >= 512 {
		for i := 0; i < 64; i++ {
			state.buffer[i] = input[i+int(index)]
		}
		f8(state)
		index, dataBitLen = index+64, dataBitLen-512
	}

	/* Store the partial block into buffer, assume that if part of the
	   last byte is not part of the message, then that part consists
	   of zero bits */
	if dataBitLen > 0 {
		plus := 0
		if (dataBitLen & 7) != 0 {
			plus = 1
		}

		/* Copy (dataBitLen & 0x1ff >> 3 + plus) bytes from input to
		   state.buffer, at an offset of index from input */
		for i := 0; i < int(((dataBitLen&0x1ff)>>3))+plus; i++ {
			state.buffer[i] = input[i+int(index)]
		}
		state.dataInBuffer = dataBitLen
	}
}

func final(state *hashState) []byte {
	/*
		Pad the message when dataBitLen is a multiple of 512 bits,
		then process the padded block
	*/
	if (state.dataBitLen & 0xff) == 0 {
		finalizeBuffer(state, true)
	} else {
		index := int(state.dataBitLen&0x1ff) >> 3
		offset := index
		if (state.dataInBuffer & 7) != 0 {
			offset++
		}

		// Set the rest of the buffer to zero
		for i := 0; i < len(state.buffer)-offset; i++ {
			state.buffer[i+offset] = 0
		}

		/*
			Pad and process the partial block when databitlen is not
			a multiple of 512 bits, then hash the padded blocks
		*/
		state.buffer[index] |= byte((1 << uint(7-(state.dataBitLen&7))))

		f8(state)

		finalizeBuffer(state, false)
	}

	output := make([]byte, 32)

	for i := 0; i < 32; i++ {
		output[i] = state.H[i+96]
	}

	return output
}

func finalizeBuffer(state *hashState, zeroInitial bool) {
	// Zero buffer
	for i := 0; i < len(state.buffer); i++ {
		state.buffer[i] = 0
	}

	if zeroInitial {
		state.buffer[0] = 0x80
	}

	for i := 0; i < 8; i++ {
		state.buffer[63-i] = byte(((state.dataBitLen >> uint(i*8)) & 0xff))
	}

	f8(state)
}

// Compression function F8
func f8(state *hashState) {
	// XOR the message with the first half of H
	for i := 0; i < 64; i++ {
		state.H[i] ^= state.buffer[i]
	}

	// Bijective function E8
	e8(state)

	// XOR the message with the last half of H
	for i := 0; i < 64; i++ {
		state.H[i+64] ^= state.buffer[i]
	}
}

func e8(state *hashState) {
	// Initialize the round constant
	for i := 0; i < 64; i++ {
		state.roundConstant[i] = roundConstantZero[i]
	}

	/*
		Initial group at the beginning of E8, group the H value into
		4-bit elements and store them in A
	*/
	e8InitialGroup(state)

	// 42 Rounds
	for i := 0; i < 42; i++ {
		r8(state)
		updateRoundConstant(state)
	}

	/*
		Degroup at the end of E8: decompose the 4-bit elements of A into
		the 1024 bit H
	*/
	e8FinalDegroup(state)
}

func e8InitialGroup(state *hashState) {
	temp := make([]byte, 256)

	for i := 0; i < 256; i++ {
		// t0 is the i-th bit of H, i = 0, 1, 2, 3, ... , 127
		t0 := byte(((state.H[(i)>>3] >> uint(7-(i&7))) & 1))

		// t1 is the (i+256)-th bit of H
		t1 := byte(((state.H[(i+256)>>3] >> uint(7-(i&7))) & 1))

		// t2 is the (i+512)-th bit of H
		t2 := byte(((state.H[(i+512)>>3] >> uint(7-(i&7))) & 1))

		// t3 is the (i+768)-th bit of H
		t3 := byte(((state.H[(i+768)>>3] >> uint(7-(i&7))) & 1))

		temp[i] = byte(((t0 << 3) | (t1 << 2) | (t2 << 1) | (t3 << 0)))
	}

	// Padding the odd and even elements separately
	for i := 0; i < 128; i++ {
		state.A[i<<1] = temp[i]
		state.A[(i<<1)+1] = temp[i+128]
	}
}

func r8(state *hashState) {
	temp := make([]byte, 256)

	// The round constant expanded into 256 1-bit elements
	roundConstantExpanded := make([]byte, 256)

	// Expand the round constant into 256 1-bit elements
	for i := 0; i < 256; i++ {
		roundConstantExpanded[i] = byte(((state.roundConstant[i>>2] >> uint(3-(i&3))) & 1))
	}

	// Sbox layer, each constant bit selects one Sbox from S0 and S1
	for i := 0; i < 256; i++ {
		// Constant bits are used to determine which Sbox to use
		temp[i] = s[roundConstantExpanded[i]][state.A[i]]
	}

	// MDS Layer
	for i := 0; i < 256; i += 2 {
		l(&temp[i], &temp[i+1])
	}

	// The following is the permutation layer P_8

	// Initial swap Pi_8
	for i := 0; i < 256; i += 4 {
		temp[i+2], temp[i+3] = temp[i+3], temp[i+2]
	}

	// Permutation P'_8
	for i := 0; i < 128; i++ {
		state.A[i] = temp[i<<1]
		state.A[i+128] = temp[(i<<1)+1]
	}

	// Final swap Phi_8
	for i := 128; i < 256; i += 2 {
		state.A[i], state.A[i+1] = state.A[i+1], state.A[i]
	}
}

func updateRoundConstant(state *hashState) {
	temp := make([]byte, 64)

	// sBox layer
	for i := 0; i < 64; i++ {
		temp[i] = s[0][state.roundConstant[i]]
	}

	// md5 layer
	for i := 0; i < 64; i += 2 {
		l(&temp[i], &temp[i+1])
	}

	// The following is the permutation layer P_6

	// Initial swap Pi_6
	for i := 0; i < 64; i += 4 {
		temp[i+2], temp[i+3] = temp[i+3], temp[i+2]
	}

	// Permutation P'_6
	for i := 0; i < 32; i++ {
		state.roundConstant[i] = temp[i<<1]
		state.roundConstant[i+32] = temp[(i<<1)+1]
	}

	// Final swap Phi_6
	for i := 32; i < 64; i += 2 {
		state.roundConstant[i], state.roundConstant[i+1] = state.roundConstant[i+1], state.roundConstant[i]
	}
}

func l(a, b *byte) {
	*b ^= byte(((*a << 1) ^ (*a >> 3) ^ ((*a >> 2) & 2)) & 0xf)
	*a ^= byte(((*b << 1) ^ (*b >> 3) ^ ((*b >> 2) & 2)) & 0xf)
}

/*
Degroup at the end of E8: it is the inverse of E8InitialGroup.
The 256 4-bit elements in state.A are degrouped into the 1024-bit
state.H
*/
func e8FinalDegroup(state *hashState) {
	temp := make([]byte, 256)
	for i := 0; i < 128; i++ {
		temp[i] = state.A[i<<1]
		temp[i+128] = state.A[(i<<1)+1]
	}

	// Zero out the array H
	for i := 0; i < len(state.H); i++ {
		state.H[i] = 0
	}

	for i := 0; i < 256; i++ {
		t0 := byte((temp[i] >> 3) & 1)
		t1 := byte((temp[i] >> 2) & 1)
		t2 := byte((temp[i] >> 1) & 1)
		t3 := byte((temp[i] >> 0) & 1)

		state.H[i>>3] |= byte(t0 << uint(7-(i&7)))
		state.H[(i+256)>>3] |= byte(t1 << uint(7-(i&7)))
		state.H[(i+512)>>3] |= byte(t2 << uint(7-(i&7)))
		state.H[(i+768)>>3] |= byte(t3 << uint(7-(i&7)))
	}
}
