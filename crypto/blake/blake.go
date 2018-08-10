/*
  BlakeSharp - Blake256
  Public domain implementation of the BLAKE hash algorithm
  by Dominik Reichl <dominik.reichl@t-online.de>
  Web: http://www.dominik-reichl.de/
  If you're using this class, it would be nice if you'd mention
  me somewhere in the documentation of your program, but it's
  not required.
  BLAKE was designed by Jean-Philippe Aumasson, Luca Henzen,
  Willi Meier and Raphael C.-W. Phan.
  BlakeSharp was derived from the reference C implementation.
  - 2018-07-04
  Modified by The TurtleCoin Developers to use 14 rounds instead of 8
  Version 1.0 - 2011-11-20
  - Initial release (implementing BLAKE v1.4).

  Modifications by BlueDragon747 for BlakeCoin Project

  Version 1.1 - 14-03-2014
  - 8 Rounds version
*/

package blake

var mT uint64

var mNBufLen int

var mBNullT bool

var mH = make([]uint, 8)

var mS = make([]uint, 4)

var mBuf = make([]byte, 64)

var mV = make([]uint, 16)

var mM = make([]uint, 16)

const nbRounds = 8

var gSigma = [nbRounds * 16]int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3,
	11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4,
	7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8,
	9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13,
	2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9,
	12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11,
	13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10}

var gCst = [16]uint{
	0x243F6A88, 0x85A308D3, 0x13198A2E, 0x03707344,
	0xA4093822, 0x299F31D0, 0x082EFA98, 0xEC4E6C89,
	0x452821E6, 0x38D01377, 0xBE5466CF, 0x34E90C6C,
	0xC0AC29B7, 0xC97C50DD, 0x3F84D5B5, 0xB5470917}

var gPadding = []byte{
	0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

const hashSizeValue = 256

func bytesToUint32(pb []byte, iOffset int) uint {
	return (uint(pb[iOffset+3]) | (uint(pb[iOffset+2]) << 8) | (uint(pb[iOffset+1]) << 16) | (uint(pb[iOffset]) << 24))
}

func uint32ToBytes(u uint, pbOut []byte, iOffset int) {
	for i := 3; i >= 0; i-- {
		pbOut[iOffset+i] = byte(u & 0xFF)
		u >>= 8
	}
}

func rotateRight(u uint, nBits int) uint {
	return (u >> uint(nBits)) | (u << (32 - uint(nBits)))
}

func g(a, b, c, d, r, i int) {
	p := (r << 4) + i
	p0 := gSigma[p]
	p1 := gSigma[p+1]

	mV[a] += mV[b] + (mM[p0] ^ gCst[p1])
	mV[d] = rotateRight(mV[d]^mV[a], 16)
	mV[c] += mV[d]
	mV[b] = rotateRight(mV[b]^mV[c], 12)
	mV[a] += mV[b] + (mM[p1] ^ gCst[p0])
	mV[d] = rotateRight(mV[d]^mV[a], 8)
	mV[c] += mV[d]
	mV[b] = rotateRight(mV[b]^mV[c], 7)
}

func compress(pbBlock []byte, iOffset int) {
	for i := 0; i < 16; i++ {
		mM[i] = bytesToUint32(pbBlock, iOffset+(i<<2))
	}

	for i := 0; i < 8; i++ {
		mV[i] = mH[i]
	}

	mV[8] = mS[0] ^ 0x243F6A88
	mV[9] = mS[1] ^ 0x85A308D3
	mV[10] = mS[2] ^ 0x13198A2E
	mV[11] = mS[3] ^ 0x03707344
	mV[12] = 0xA4093822
	mV[13] = 0x299F31D0
	mV[14] = 0x082EFA98
	mV[15] = 0xEC4E6C89

	if !mBNullT {
		uLen := uint(mT & 0xFFFFFFFF)
		mV[12] ^= uLen
		mV[13] ^= uLen
		uLen = uint((mT >> 32) & 0xFFFFFFFF)
		mV[14] ^= uLen
		mV[15] ^= uLen
	}

	for r := 0; r < nbRounds; r++ {
		g(0, 4, 8, 12, r, 0)
		g(1, 5, 9, 13, r, 2)
		g(2, 6, 10, 14, r, 4)
		g(3, 7, 11, 15, r, 6)
		g(3, 4, 9, 14, r, 14)
		g(2, 7, 8, 13, r, 12)
		g(0, 5, 10, 15, r, 8)
		g(1, 6, 11, 12, r, 10)
	}

	for i := 0; i < 8; i++ {
		mH[i] ^= mV[i]
	}

	for i := 0; i < 8; i++ {
		mH[i] ^= mV[i+8]
	}

	for i := 0; i < 4; i++ {
		mH[i] ^= mS[i]
	}

	for i := 0; i < 4; i++ {
		mH[i+4] ^= mS[i]
	}
}

func hashCore(array []byte, ibStart, cbSize int) {
	iOffset := ibStart
	nFill := 64 - mNBufLen

	if mNBufLen > 0 && cbSize >= nFill {

		for i := 0; i < nFill; i++ {
			mBuf[mNBufLen+i] = array[iOffset+i]
		}

		mT += 512
		compress(mBuf, 0)
		iOffset += nFill
		cbSize -= nFill
		mNBufLen = 0
	}

	for cbSize >= 64 {
		mT += 512
		compress(array, iOffset)
		iOffset += 64
		cbSize -= 64
	}

	if cbSize > 0 {
		for i := 0; i < cbSize; i++ {
			mBuf[i+mNBufLen] = array[i+iOffset]
		}
		mNBufLen += cbSize
	} else {
		mNBufLen = 0
	}
}

func hashFinal() []byte {
	pbMsgLen := make([]byte, 8)
	uLen := mT + (uint64(mNBufLen) << 3)
	uint32ToBytes(uint((uLen>>32)&0xFFFFFFFF), pbMsgLen, 0)
	uint32ToBytes(uint(uLen&0xFFFFFFFF), pbMsgLen, 4)

	if mNBufLen == 55 {
		mT -= 8
		hashCore([]byte{0x81}, 0, 1)
	} else {
		if mNBufLen < 55 {
			if mNBufLen == 0 {
				mBNullT = true
			}
			mT -= uint64(440) - (uint64(mNBufLen) << 3)
			hashCore(gPadding, 0, 55-mNBufLen)
		} else {
			mT -= uint64(512) - (uint64(mNBufLen) << 3)
			hashCore(gPadding, 0, 64-mNBufLen)
			mT -= uint64(440)
			hashCore(gPadding, 1, 55)
			mBNullT = true
		}
		hashCore([]byte{0x01}, 0, 1)
		mT -= 8
	}
	mT -= 64
	hashCore(pbMsgLen, 0, 8)

	pbDigest := make([]byte, 32)

	for i := 0; i < 8; i++ {
		uint32ToBytes(mH[i], pbDigest, i<<2)
	}

	return pbDigest
}

// ComputeHash calculates the blake256 hash
// of corresponding input and returns it.
func ComputeHash(input []byte) []byte {
	mH[0] = 0x6A09E667
	mH[1] = 0xBB67AE85
	mH[2] = 0x3C6EF372
	mH[3] = 0xA54FF53A
	mH[4] = 0x510E527F
	mH[5] = 0x9B05688C
	mH[6] = 0x1F83D9AB
	mH[7] = 0x5BE0CD19

	for i := 0; i < len(mS); i++ {
		mS[i] = 0
	}

	mT = 0
	mNBufLen = 0
	mBNullT = false

	for i := 0; i < len(mBuf); i++ {
		mBuf[i] = 0
	}

	hashCore(input, 0, len(input))
	hashValue := hashFinal()

	return hashValue
}
