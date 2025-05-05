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
	csVariablesMapping = map[CastlingSide][5]int{
		WHITE_KING_SIDE:  {0, 2, 3, 1, 3},
		WHITE_QUEEN_SIDE: {7, 4, 3, 5, 3},
		BLACK_KING_SIDE:  {56, 58, 59, 57, 12},
		BLACK_QUEEN_SIDE: {63, 60, 59, 61, 12},
	}

	//Mapping from string representation to CastlingSide type
	csMoveMapping = map[string]CastlingSide{
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
		ClearBit(csVariablesMapping[castlingType][0], &chess.WhiteRooks)
		SetBit(csVariablesMapping[castlingType][1], &chess.WhiteRooks)
		ClearBit(csVariablesMapping[castlingType][2], &chess.WhiteKing)
		SetBit(csVariablesMapping[castlingType][3], &chess.WhiteKing)
	} else {
		ClearBit(csVariablesMapping[castlingType][0], &chess.BlackRooks)
		SetBit(csVariablesMapping[castlingType][1], &chess.BlackRooks)
		ClearBit(csVariablesMapping[castlingType][2], &chess.BlackKing)
		SetBit(csVariablesMapping[castlingType][3], &chess.BlackKing)

		//Only Black turn that full move can increase
		chess.Fullmove++
	}

	chess.SideToMove = !chess.SideToMove
	chess.CastlingPrivilege &= csVariablesMapping[castlingType][4]
}

// Makemove method. Here, we assume that the move is a valid move: correct move syntax, correct turn and valid move
func (chess *Chess) MakeMove(move string) {
	//If move is castling, delegate it to castling funtion
	if csType, ok := csMoveMapping[move]; ok {
		chess.Castling(csType)
		return
	}

	//If other moves, we calculate sourceIndex, destIndex and captureIndex
	var (
		sourceIndex  = FromAlgebraicToIndex(move[:2])
		destIndex    = FromAlgebraicToIndex(move[2:4])
		captureIndex int
	)

	//Reset en passant target
	chess.EnPassantTarget = -1

	//Moving the piece
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
		//Find the bitboard of the piece that moving
		switch {
		case IsPieceAtIndex(chess.WhitePawns, sourceIndex):
			//If this is pawn double push, re-calculate en passant target
			if Abs(destIndex-sourceIndex) == 16 {
				chess.EnPassantTarget = (destIndex + sourceIndex) / 2
			}
			//Place the pawn at destIndex
			SetBit(destIndex, &chess.WhitePawns)
			//Reset halfmove counter to 0 (50 moves rule) when a pawn moving
			chess.Halfmove = 0
		case IsPieceAtIndex(chess.WhiteRooks, sourceIndex):
			SetBit(destIndex, &chess.WhiteRooks)
		case IsPieceAtIndex(chess.WhiteKnights, sourceIndex):
			SetBit(destIndex, &chess.WhiteKnights)
		case IsPieceAtIndex(chess.WhiteBishops, sourceIndex):
			SetBit(destIndex, &chess.WhiteBishops)
		case IsPieceAtIndex(chess.WhiteQueens, sourceIndex):
			SetBit(destIndex, &chess.WhiteQueens)
		case IsPieceAtIndex(chess.WhiteKing, sourceIndex):
			SetBit(destIndex, &chess.WhiteKing)
		case IsPieceAtIndex(chess.BlackPawns, sourceIndex):
			//If this is pawn double push, re-calculate en passant target
			if Abs(destIndex-sourceIndex) == 16 {
				chess.EnPassantTarget = (destIndex + sourceIndex) / 2
			}
			//Place the pawn at destIndex
			SetBit(destIndex, &chess.BlackPawns)
			//Reset halfmove counter to 0 (50 moves rule) when a pawn moving
			chess.Halfmove = 0
		case IsPieceAtIndex(chess.BlackRooks, sourceIndex):
			SetBit(destIndex, &chess.BlackRooks)
		case IsPieceAtIndex(chess.BlackKnights, sourceIndex):
			SetBit(destIndex, &chess.BlackKnights)
		case IsPieceAtIndex(chess.BlackBishops, sourceIndex):
			SetBit(destIndex, &chess.BlackBishops)
		case IsPieceAtIndex(chess.BlackQueens, sourceIndex):
			SetBit(destIndex, &chess.BlackQueens)
		case IsPieceAtIndex(chess.BlackKing, sourceIndex):
			SetBit(destIndex, &chess.BlackKing)
		}
	}

	//Move the piece
	if chess.SideToMove {
		//Remove the piece stay at sourceIndex
		ClearBitAcrossBoards(sourceIndex, &chess.WhitePawns, &chess.WhiteRooks, &chess.WhiteKnights,
			&chess.WhiteBishops, &chess.WhiteQueens, &chess.WhiteKing)

		//Reset the halfmove counter to 0 (50 moves rule) if there is a capture
		if IsPieceAtIndex(chess.GenerateAllBlacks(), captureIndex) {
			chess.Halfmove = 0
		}

		//Calculate capture index
		if strings.HasSuffix(move, "EP") { //If this is en passant move
			captureIndex = destIndex - 8
		} else {
			captureIndex = destIndex
		}

		//Remove capture piece
		ClearBitAcrossBoards(captureIndex, &chess.BlackPawns, &chess.BlackRooks, &chess.BlackKnights,
			&chess.BlackBishops, &chess.BlackQueens)
	} else {
		//Remove the piece stay at sourceIndex
		ClearBitAcrossBoards(sourceIndex, &chess.BlackPawns, &chess.BlackRooks, &chess.BlackKnights,
			&chess.BlackBishops, &chess.BlackQueens, &chess.BlackKing)

		//Reset the halfmove counter to 0 (50 moves rule) if there is a capture
		if IsPieceAtIndex(chess.GenerateAllWhites(), captureIndex) {
			chess.Halfmove = 0
		}

		//Calculate capture index
		if strings.HasSuffix(move, "EP") { //If this is en passant move
			captureIndex = destIndex + 8
		} else {
			captureIndex = destIndex
		}

		//Remove capture piece
		ClearBitAcrossBoards(captureIndex, &chess.WhitePawns, &chess.WhiteRooks, &chess.WhiteKnights,
			&chess.WhiteBishops, &chess.WhiteQueens)

		//Update fullmove (can only increase in Black turn)
		chess.Fullmove++
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

	chess.SideToMove = !chess.SideToMove
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

	return result, total
}
