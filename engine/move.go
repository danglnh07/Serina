package engine

import (
	"fmt"
	"strings"
	"sync"
)

// Mapping for castling (to avoid if else)
var (

	/*
	 * Mapping for the castling operation. Each element in the array are:
	 * Initial rook position
	 * New rook position
	 * Initial king position
	 * New king position
	 * Castling turn off bitmask
	 */
	csMap = map[CastlingSide][5]int{
		WHITE_KING_SIDE:  {0, 2, 3, 1, 3},
		WHITE_QUEEN_SIDE: {7, 4, 3, 5, 3},
		BLACK_KING_SIDE:  {56, 58, 59, 57, 12},
		BLACK_QUEEN_SIDE: {63, 60, 59, 61, 12},
	}

	//Mapping from string representation to CastlingSide type
	csDecision = map[string]CastlingSide{
		"O-O":   WHITE_KING_SIDE,
		"O-O-O": WHITE_QUEEN_SIDE,
		"o-o":   BLACK_KING_SIDE,
		"o-o-o": BLACK_QUEEN_SIDE,
	}
)

// Method to perform castling
func (chess *Chess) Castling(castlingType CastlingSide) {
	//Set en passant target, full move and half move
	chess.EnPassantTarget = -1
	chess.Halfmove++

	if castlingType == WHITE_KING_SIDE || castlingType == WHITE_QUEEN_SIDE {
		ClearBit(csMap[castlingType][0], &chess.WhiteRooks)
		SetBit(csMap[castlingType][1], &chess.WhiteRooks)
		ClearBit(csMap[castlingType][2], &chess.WhiteKing)
		SetBit(csMap[castlingType][3], &chess.WhiteKing)
	} else {
		ClearBit(csMap[castlingType][0], &chess.BlackRooks)
		SetBit(csMap[castlingType][1], &chess.BlackRooks)
		ClearBit(csMap[castlingType][2], &chess.BlackKing)
		SetBit(csMap[castlingType][3], &chess.BlackKing)

		//Only Black turn that full move can increase
		chess.Fullmove++
	}

	chess.SideToMove = !chess.SideToMove
	chess.CastlingPrivilege &= csMap[castlingType][4]
}

// Makemove method, which will modify the chess struct directly
func (chess *Chess) MakeMove(move string) {
	//If move is castling, delegate it to castling funtion
	if csType, ok := csDecision[move]; ok {
		//Check if the side to move is correct
		if ((csType == WHITE_KING_SIDE || csType == WHITE_QUEEN_SIDE) && !chess.SideToMove) ||
			((csType == BLACK_KING_SIDE || csType == BLACK_QUEEN_SIDE) && chess.SideToMove) {
			fmt.Println("Not your turn!")
			return
		}

		chess.Castling(csType)
		return
	}

	//If other moves, we calculate sourceIndex, destIndex and captureIndex
	var (
		sourceIndex  = FromAlgebraicToIndex(move[:2])
		destIndex    = FromAlgebraicToIndex(move[2:4])
		captureIndex int
		//Hash map for traversal
		movedMap map[uint64]*uint64
	)

	//We first check if the turn is correct
	if (chess.SideToMove && IsPieceAtIndex(chess.GenerateAllBlacks(), sourceIndex)) || (!chess.SideToMove && IsPieceAtIndex(chess.GenerateAllWhites(), sourceIndex)) {
		fmt.Println("Not your turn!")
		return
	}

	//Calculate capture index
	if strings.HasSuffix(move, "EP") { //If this is en passant move
		if chess.SideToMove { //If WHITE turn
			captureIndex = destIndex - 8
		} else { //If BLACK turn
			captureIndex = destIndex + 8
		}
	} else {
		captureIndex = destIndex
	}

	//Calculate maps for traversal
	if chess.SideToMove {
		movedMap = map[uint64]*uint64{
			chess.WhitePawns:   &chess.WhitePawns,
			chess.WhiteRooks:   &chess.WhiteRooks,
			chess.WhiteKnights: &chess.WhiteKnights,
			chess.WhiteBishops: &chess.WhiteBishops,
			chess.WhiteQueens:  &chess.WhiteQueens,
			chess.WhiteKing:    &chess.WhiteKing,
		}
	} else {
		movedMap = map[uint64]*uint64{
			chess.BlackPawns:   &chess.BlackPawns,
			chess.BlackRooks:   &chess.BlackRooks,
			chess.BlackKnights: &chess.BlackKnights,
			chess.BlackBishops: &chess.BlackBishops,
			chess.BlackQueens:  &chess.BlackQueens,
			chess.BlackKing:    &chess.BlackKing,
		}
	}

	//Move the piece
	for key, val := range movedMap {
		if IsPieceAtIndex(key, sourceIndex) {
			//If there is a pawn double push, re-calculate en passant target, else reset it to -1
			if (IsPieceAtIndex(chess.WhitePawns, sourceIndex) || IsPieceAtIndex(chess.BlackPawns, sourceIndex)) && Abs(destIndex-sourceIndex) == 16 {
				chess.EnPassantTarget = (destIndex + sourceIndex) / 2
			} else {
				chess.EnPassantTarget = -1
			}

			ClearBit(sourceIndex, val)

			//Check if there is a promotion, if not then we set the bit at destIndex
			switch move[len(move)-1] {
			case 'R':
				SetBit(destIndex, &chess.WhiteRooks)
			case 'N':
				SetBit(destIndex, &chess.WhiteKnights)
			case 'B':
				SetBit(destIndex, &chess.WhiteBishops)
			case 'Q':
				SetBit(destIndex, &chess.WhiteQueens)
			case 'r':
				SetBit(destIndex, &chess.BlackRooks)
			case 'n':
				SetBit(destIndex, &chess.BlackKnights)
			case 'b':
				SetBit(destIndex, &chess.BlackBishops)
			case 'q':
				SetBit(destIndex, &chess.BlackQueens)
			default:
				SetBit(destIndex, val)
			}

			//If found, no need to loop more
			break
		}
	}

	//Remove capture piece
	if chess.SideToMove {
		ClearBitAcrossBoards(captureIndex, &chess.BlackPawns, &chess.BlackRooks, &chess.BlackKnights, &chess.BlackBishops, &chess.BlackQueens)
	} else {
		ClearBitAcrossBoards(captureIndex, &chess.WhitePawns, &chess.WhiteRooks, &chess.WhiteKnights, &chess.WhiteBishops, &chess.WhiteQueens)
	}

	//Re-calculate castling privilege
	if !IsPieceAtIndex(chess.WhiteKing, 3) {
		chess.CastlingPrivilege &= 3
	}

	if !IsPieceAtIndex(chess.BlackKing, 59) {
		chess.CastlingPrivilege &= 12
	}

	if !IsPieceAtIndex(chess.WhiteRooks, 0) {
		chess.CastlingPrivilege &= 7
	}

	if !IsPieceAtIndex(chess.WhiteRooks, 7) {
		chess.CastlingPrivilege &= 11
	}

	if !IsPieceAtIndex(chess.BlackRooks, 56) {
		chess.CastlingPrivilege &= 13
	}

	if !IsPieceAtIndex(chess.BlackRooks, 63) {
		chess.CastlingPrivilege &= 14
	}

	//Update sideToMove, halfmove and fullmove
	if !chess.SideToMove {
		chess.Fullmove++
	}
	chess.SideToMove = !chess.SideToMove
	chess.Halfmove++
}

// Perft method: return all move found in the n-depthed search tree. We use copy here instead of Make/Unmake. This is a single threaded version
func (chess *Chess) Perft(depth int) int {
	if depth == 0 {
		return 1
	}

	count := 0
	moves := chess.MoveGeneration()
	for _, move := range moves {
		clone := chess.Clone()
		clone.MakeMove(move)
		count += clone.Perft(depth - 1)
	}

	return count
}

// Method to print the divide perft result both to the standard output and to a text file for further debugging.
// This method use the Perft function, and also single threaded
// It will return a map which is the divide perft result, and total node found
func (chess *Chess) DividePerft(depth int) (map[string]int, int) {
	//Return nil and 0 if the we reach the final depth
	if depth <= 0 {
		return nil, 1
	}

	//Variables declaration
	results := make(map[string]int)
	moves := chess.MoveGeneration()
	total := 0

	//For each move, we clone the chess, perform the move and run the Perft function at depth - 1
	for _, move := range moves {
		clone := chess.Clone()
		clone.MakeMove(move)
		count := clone.Perft(depth - 1)
		total += count
		results[move] = count
	}

	//Turn the map to string for printing out
	str := ""
	for move, count := range results {
		str += fmt.Sprintf("%s: %d\n", move, count)
	}
	WriteFile(str, "perft_test.txt") //Write to perft_test.txt file
	str += fmt.Sprintf("Node found: %d\n", total)
	fmt.Println(str)

	return results, total
}

// Method to print the divide perft, but use goroutine
func (chess *Chess) FastPerft(depth int) (map[string]int, int) {
	//If a depth is small (which means the total node would also not large), we use the single threaded version to not watse resources
	if depth <= 3 {
		return chess.DividePerft(depth)
	}

	//Variables declaration
	var (
		wg     sync.WaitGroup
		mutex  sync.Mutex
		total                 = 0
		result map[string]int = make(map[string]int)
	)

	//Generate all moves, and loop through each node
	moves := chess.MoveGeneration()
	for _, move := range moves {
		//We add 1 job to the wait group
		wg.Add(1)

		//Spawn goroutine for only the direct child of the root
		go func(moveStr string) {
			//Signify job done to decrease the wait group
			defer wg.Done()

			//Clone and make move
			clone := chess.Clone()
			clone.MakeMove(moveStr)

			//Run perft for the smaller tree
			count := clone.Perft(depth - 1)

			//Add move and node found to the result map. We use mutex to avoid race condition since result is a share resource
			mutex.Lock()
			result[move] = count
			total += count
			mutex.Unlock()
		}(move) //Pass move as a parameter to avoid closure issues
	}

	//Wait for all goroutine to done their job before continue
	wg.Wait()

	//Print the divide perft
	str := ""
	for move, count := range result {
		str += fmt.Sprintf("%s: %d\n", move, count)
	}
	WriteFile(str, "perft_test.txt")
	str += fmt.Sprintf("Node found: %d", total)
	fmt.Println(str)

	return result, total
}
