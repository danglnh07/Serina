package engine

import (
	"fmt"
	"os"
	"strings"
)

// Method to convert from algebraic to index
// Index 0 is H1, and 63 is A8
// For convention in this engine, all File (A-H) will be uppercase
func FromAlgebraicToIndex(algebraic string) int {
	//Set default to uppercase
	algebraic = strings.ToUpper(algebraic)
	return (8 * int(algebraic[1]-'0')) - int(algebraic[0]-'A') - 1
}

// Method for convert index back to algebraic notation
func FromIndexToAlgebraic(index int) string {
	/*
	 * Rank: (index div 8) + 1
	 * File: 'A' + 7 - (index mod 8) = 'H' - (index mod 8)
	 */
	return fmt.Sprintf("%c%d", 'H'-rune(index%8), 1+index/8)
}

// Method to check if a there is a piece at a certain index in a bitboard
// If index is invalid (not in range [0, 63]) then it would be false by default
func IsPieceAtIndex(bitboard uint64, index int) bool {
	return index >= 0 && index <= 63 && ((bitboard>>index)&0x1 == 0x1)
}

// Method for setting the bit at an index to 0 (the bitboard is manipulated directly)
func ClearBit(index int, bitboard *uint64) {
	*bitboard &= ^(0x1 << index)
}

// Method for setting the bit of multiple bitboards at same index to 0 (the bitboard is manipulated directly)
func ClearBitAcrossBoards(index int, bitboards ...*uint64) {
	for _, val := range bitboards {
		ClearBit(index, val)
	}
}

// Method for setting the bit at an index to 1 (the bitboard is manipulated directly)
func SetBit(index int, bitboard *uint64) {
	*bitboard |= 0x1 << index
}

// Method for setting the bit of multiple bitboards at same index to 1 (the bitboard is manipulated directly)
func SetBitAcrossBoards(index int, bitboards ...*uint64) {
	for _, val := range bitboards {
		SetBit(index, val)
	}
}

// Method to check if two square index is at same Rank
func IsAtSameRank(index1 int, index2 int) bool {
	return index1/8 == index2/8
}

// Method to check if two square index is at same File
func IsAtSameFile(index1 int, index2 int) bool {
	return index1%8 == index2%8
}

// Method to check if two square index is at same Diagonal
func IsAtSameDiagonal(index1 int, index2 int) bool {
	return index1/8+index1%8 == index2/8+index2%8
}

// Method to check if two square index is at same Anti Diagonal
func IsAtSameAntiDiagonal(index1 int, index2 int) bool {
	return index1/8-index1%8 == index2/8-index2%8
}

// Return the bitboard of all squares between min and max in a dicrection (excluding)
func CalculateRayAttackLine(min, max int, direction Direction) uint64 {
	var rayline uint64 = 0
	for i := min + int(direction); i < max; i += int(direction) {
		rayline += 0x1 << i
	}
	return rayline
}

// Method for printing an unsigned integer to 8x8 matrix of bits
// The LSB is at the bottom right of the matrix
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

// Get the smaller value between 2 integers
func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// Get the bigger value between 2 integers
func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// Get the absolute value between 2 integers
func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Method for writing string data into a file. Useful for debugging perft
func WriteFile(str, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	str = strings.ToLower(str)

	_, err = file.WriteString(str)
	if err != nil {
		return err
	}

	return nil
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

// Flip index of the chessboard's square vertically
func FlipIndexVertical(index int) int {
	return 8*(7-index/8) + index%8
}
