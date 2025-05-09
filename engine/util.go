package engine

import (
	"fmt"
	"strings"
)

func FromAlgebraicToIndex(algebraic string) int {
	algebraic = strings.ToLower(algebraic)
	return (8 * int(algebraic[1]-'0')) - int(algebraic[0]-'a') - 1
}

func FromIndexToAlgebraic(index int) string {
	/*
	 * Rank: (index div 8) + 1
	 * File: 'A' + 7 - (index mod 8) = 'H' - (index mod 8)
	 */
	return fmt.Sprintf("%c%d", 'h'-rune(index%8), 1+index/8)
}

func IsPieceAtIndex(bitboard uint64, index int) bool {
	return index >= 0 && index <= 63 && ((bitboard>>index)&0x1 == 0x1)
}

func ClearBit(index int, bitboard *uint64) {
	*bitboard &= ^(0x1 << index)
}

func ClearBitAcrossBoards(index int, bitboards ...*uint64) {
	for _, val := range bitboards {
		ClearBit(index, val)
	}
}

func SetBit(index int, bitboard *uint64) {
	*bitboard |= 0x1 << index
}

func SetBitAcrossBoards(index int, bitboards ...*uint64) {
	for _, val := range bitboards {
		SetBit(index, val)
	}
}

func IsAtSameRank(index1 int, index2 int) bool {
	return index1/8 == index2/8
}

func IsAtSameFile(index1 int, index2 int) bool {
	return index1%8 == index2%8
}

func IsAtSameDiagonal(index1 int, index2 int) bool {
	return index1/8+index1%8 == index2/8+index2%8
}

func IsAtSameAntiDiagonal(index1 int, index2 int) bool {
	return index1/8-index1%8 == index2/8-index2%8
}

func CalculateRayAttackLine(min, max int, direction int) uint64 {
	var rayline uint64
	switch direction {
	case RANK:
		rayline = RANK_MASK[min/8]
	case FILE:
		rayline = FILE_MASK[7-min%8]
	case DIAGONAL:
		rayline = DIAGONAL_MASK[14-(min/8+min%8)]
	case ANTI_DIAGONAL:
		rayline = ANTI_DIAGONAL_MASK[7-(min/8-min%8)]
	}
	return rayline & ((0x1 << max) - 1) & ^((0x1 << (min + 1)) - 1)
}

func PrintBitboard(bitboard uint64) {
	var str string = ""
	for i := 0; i < 64; i++ {
		str = fmt.Sprintf("%d", bitboard&0x1) + str
		if i%8 == 7 {
			str = "\n" + str
		}
		bitboard >>= 0x1
	}
	fmt.Println(str)
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func FlipVertical(bitboard uint64) uint64 {
	return (bitboard << 56) | //Map from RANK 1 to RANK 8
		((bitboard << 40) & RANK_MASK[6]) | //Map from RANK 2 to RANK 7
		((bitboard << 24) & RANK_MASK[5]) | //Map from RANK 3 to RANK 6
		((bitboard << 8) & RANK_MASK[4]) | //Map from RANK 4 to RANK 5
		((bitboard >> 8) & RANK_MASK[3]) | //Map from RANK 5 to RANK 4
		((bitboard >> 24) & RANK_MASK[2]) | //Map from RANK 6 to RANK 3
		((bitboard >> 40) & RANK_MASK[1]) | //Map from RANK 7 to RANK 2
		(bitboard >> 56) //Map from RANK 8 to RANK 1
}

func FlipIndexVertical(index int) int {
	return 8*(7-index/8) + index%8
}
