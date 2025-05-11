package engine

import (
	"fmt"
	"sync"
)

type Move struct {
	Castling  int
	FromBoard int
	FromIndex int
	ToBoard   int
	ToIndex   int
}

func (move Move) String() string {
	//Empty move/invalid move
	if move.ToIndex == move.FromIndex {
		return ""
	}

	switch move.Castling {
	case WHITE_KING_SIDE:
		return "O-O"
	case WHITE_QUEEN_SIDE:
		return "O-O-O"
	case BLACK_KING_SIDE:
		return "o-o"
	case BLACK_QUEEN_SIDE:
		return "o-o-o"
	default:
		str := fmt.Sprintf("%s%s", FromIndexToAlgebraic(move.FromIndex), FromIndexToAlgebraic(move.ToIndex))
		if move.ToBoard != move.FromBoard {
			switch move.ToBoard {
			case 1:
				str += "R"
			case 2:
				str += "N"
			case 3:
				str += "B"
			case 4:
				str += "Q"
			}
		}

		return str
	}
}

func NewMove(chess *Chess, str string) Move {
	switch str {
	case "O-O":
		return Move{Castling: WHITE_KING_SIDE}
	case "O-O-O":
		return Move{Castling: WHITE_QUEEN_SIDE}
	case "o-o":
		return Move{Castling: BLACK_KING_SIDE}
	case "o-o-o":
		return Move{Castling: BLACK_QUEEN_SIDE}
	default:
		move := Move{Castling: 0}
		move.FromIndex = FromAlgebraicToIndex(str[:2])
		move.ToIndex = FromAlgebraicToIndex(str[2:4])

		switch {
		case IsPieceAtIndex(chess.Boards[WHITE_PAWN], move.FromIndex):
			move.FromBoard = WHITE_PAWN
			if len(str) == 5 {
				switch str[4] {
				case 'R':
					move.ToBoard = WHITE_ROOK
				case 'N':
					move.ToBoard = WHITE_KNIGHT
				case 'B':
					move.ToBoard = WHITE_BISHOP
				case 'Q':
					move.ToBoard = WHITE_QUEEN
				}
			} else {
				move.ToBoard = WHITE_PAWN
			}
		case IsPieceAtIndex(chess.Boards[BLACK_PAWN], move.FromIndex):
			move.FromBoard = BLACK_PAWN
			if len(str) == 5 {
				switch str[4] {
				case 'r':
					move.ToBoard = BLACK_ROOK
				case 'n':
					move.ToBoard = BLACK_KNIGHT
				case 'b':
					move.ToBoard = BLACK_BISHOP
				case 'q':
					move.ToBoard = BLACK_QUEEN
				}
			} else {
				move.ToBoard = BLACK_PAWN
			}
		case IsPieceAtIndex(chess.Boards[WHITE_ROOK], move.FromIndex):
			move.FromBoard, move.ToBoard = WHITE_ROOK, WHITE_ROOK
		case IsPieceAtIndex(chess.Boards[WHITE_KNIGHT], move.FromIndex):
			move.FromBoard, move.ToBoard = WHITE_KNIGHT, WHITE_KNIGHT
		case IsPieceAtIndex(chess.Boards[WHITE_BISHOP], move.FromIndex):
			move.FromBoard, move.ToBoard = WHITE_BISHOP, WHITE_BISHOP
		case IsPieceAtIndex(chess.Boards[WHITE_QUEEN], move.FromIndex):
			move.FromBoard, move.ToBoard = WHITE_QUEEN, WHITE_QUEEN
		case IsPieceAtIndex(chess.Boards[WHITE_KING], move.FromIndex):
			move.FromBoard, move.ToBoard = WHITE_KING, WHITE_KING
		case IsPieceAtIndex(chess.Boards[BLACK_ROOK], move.FromIndex):
			move.FromBoard, move.ToBoard = BLACK_ROOK, BLACK_ROOK
		case IsPieceAtIndex(chess.Boards[BLACK_KNIGHT], move.FromIndex):
			move.FromBoard, move.ToBoard = BLACK_KNIGHT, BLACK_KNIGHT
		case IsPieceAtIndex(chess.Boards[BLACK_BISHOP], move.FromIndex):
			move.FromBoard, move.ToBoard = BLACK_BISHOP, BLACK_BISHOP
		case IsPieceAtIndex(chess.Boards[BLACK_QUEEN], move.FromIndex):
			move.FromBoard, move.ToBoard = BLACK_QUEEN, BLACK_QUEEN
		case IsPieceAtIndex(chess.Boards[BLACK_KING], move.FromIndex):
			move.FromBoard, move.ToBoard = BLACK_KING, BLACK_KING
		}

		return move
	}
}

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
	csMapping = map[int][5]int{
		WHITE_KING_SIDE:  {0, 2, 3, 1, 3},
		WHITE_QUEEN_SIDE: {7, 4, 3, 5, 3},
		BLACK_KING_SIDE:  {56, 58, 59, 57, 12},
		BLACK_QUEEN_SIDE: {63, 60, 59, 61, 12},
	}
)

// Method to perform castling
func (chess *Chess) Castling(cs int) {
	//Set en passant target, full move and half move
	chess.EnPassantTarget = -1
	chess.Halfmove++

	if cs == WHITE_KING_SIDE || cs == WHITE_QUEEN_SIDE {
		ClearBit(csMapping[cs][0], &chess.Boards[WHITE_ROOK])
		SetBit(csMapping[cs][1], &chess.Boards[WHITE_ROOK])
		ClearBit(csMapping[cs][2], &chess.Boards[WHITE_KING])
		SetBit(csMapping[cs][3], &chess.Boards[WHITE_KING])
	} else {
		ClearBit(csMapping[cs][0], &chess.Boards[BLACK_ROOK])
		SetBit(csMapping[cs][1], &chess.Boards[BLACK_ROOK])
		ClearBit(csMapping[cs][2], &chess.Boards[BLACK_KING])
		SetBit(csMapping[cs][3], &chess.Boards[BLACK_KING])

		//Only Black turn that full move can increase
		chess.Fullmove++
	}

	chess.SideToMove = WHITE + BLACK - chess.SideToMove
	chess.CastlingPrivilege &= csMapping[cs][4]
}

// Makemove method. Here, we assume that the move is a valid move: correct move syntax, correct turn and valid move
func (chess *Chess) MakeMove(move Move) {
	if move.Castling != 0 {
		chess.Castling(move.Castling)
		return
	}

	//Move the piece
	ClearBit(move.FromIndex, &chess.Boards[move.FromBoard])

	//Place the piece down
	SetBit(move.ToIndex, &chess.Boards[move.ToBoard])

	//Calculate capture index and remove the capture piece
	captureIndex := move.ToIndex
	if chess.SideToMove == WHITE {
		if move.FromBoard == WHITE_PAWN && move.ToIndex == chess.EnPassantTarget {
			captureIndex = chess.EnPassantTarget - 8
			ClearBit(captureIndex, &chess.Boards[BLACK_PAWN])
		} else {
			for i := BLACK_PAWN; i < BLACK_KING; i++ { //King capturing normally not happen, so we ignore it here
				ClearBit(captureIndex, &chess.Boards[i])
			}
		}
	} else {
		if move.FromBoard == BLACK_PAWN && move.ToIndex == chess.EnPassantTarget {
			captureIndex = chess.EnPassantTarget + 8
			ClearBit(captureIndex, &chess.Boards[WHITE_PAWN])
		} else {
			for i := WHITE_PAWN; i < WHITE_KING; i++ {
				ClearBit(captureIndex, &chess.Boards[i])
			}
		}
	}

	//Re-calculate game state
	if (move.FromBoard == WHITE_PAWN || move.FromBoard == BLACK_PAWN) && (Abs(move.ToIndex-move.FromIndex)) == 16 {
		chess.EnPassantTarget = (move.ToIndex + move.FromIndex) / 2
	} else {
		chess.EnPassantTarget = -1
	}

	if !IsPieceAtIndex(chess.Boards[WHITE_KING], 3) {
		chess.CastlingPrivilege &= 3
	} else {
		if !IsPieceAtIndex(chess.Boards[WHITE_ROOK], 0) {
			chess.CastlingPrivilege &= 7
		}

		if !IsPieceAtIndex(chess.Boards[WHITE_ROOK], 7) {
			chess.CastlingPrivilege &= 11
		}
	}

	if !IsPieceAtIndex(chess.Boards[BLACK_KING], 59) {
		chess.CastlingPrivilege &= 12
	} else {
		if !IsPieceAtIndex(chess.Boards[BLACK_ROOK], 56) {
			chess.CastlingPrivilege &= 13
		}

		if !IsPieceAtIndex(chess.Boards[BLACK_ROOK], 63) {
			chess.CastlingPrivilege &= 14
		}
	}

	chess.SideToMove = WHITE + BLACK - chess.SideToMove
	if move.FromBoard == WHITE_PAWN || move.FromBoard == BLACK_PAWN || IsPieceAtIndex(chess.GenerateAllBlacks(), captureIndex) {
		chess.Halfmove = 0
	} else {
		chess.Halfmove++
	}

	if chess.SideToMove == BLACK {
		chess.Fullmove++
	}
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
		results[move.String()] = count
	}

	//Turn the map to string for printing out
	str := ""
	for move, count := range results {
		str += fmt.Sprintf("%s: %d\n", move, count)
	}

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
		go func(move Move) {
			//Signify job done to decrease the wait group
			defer wg.Done()

			//Clone and make move
			clone := chess.Clone()
			clone.MakeMove(move)

			//Run perft for the smaller tree
			count := clone.Perft(depth - 1)

			//Add move and node found to the result map. We use mutex to avoid race condition since result is a share resource
			mutex.Lock()
			result[move.String()] = count
			total += count
			mutex.Unlock()
		}(move) //Pass move as a parameter to avoid closure issues
	}

	//Wait for all goroutine to done their job before continue
	wg.Wait()

	return result, total
}
