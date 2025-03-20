package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Method for clearing all data in Chess struct to default value
func (chess *Chess) Clear() {
	//Clear all data in chess struct
	chess.WhitePawns = 0
	chess.WhiteRooks = 0
	chess.WhiteKnights = 0
	chess.WhiteBishops = 0
	chess.WhiteQueens = 0
	chess.WhiteKing = 0
	chess.BlackPawns = 0
	chess.BlackRooks = 0
	chess.BlackKnights = 0
	chess.BlackBishops = 0
	chess.BlackQueens = 0
	chess.BlackKing = 0
	chess.EnPassantTarget = -1
	chess.CastlingPrivilege = 0
	chess.SideToMove = true
	chess.Halfmove = 0
	chess.Fullmove = 1
}

// Method for setting up the chess board based on a FEN string
// If the string is empty, then it would be starting position
func (chess *Chess) FEN(fen string) {
	//Clear the old data of chess object (in case of replay)
	chess.Clear()

	//If empty string is provided, then we assumed it to be default position
	if len(fen) == 0 || fen == "" {
		fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" //Default value
	}

	//Split the fen string
	data := strings.Split(fen, " ")

	//Set up the bitboards
	conv := map[rune]*uint64{
		'P': &chess.WhitePawns,
		'R': &chess.WhiteRooks,
		'N': &chess.WhiteKnights,
		'B': &chess.WhiteBishops,
		'Q': &chess.WhiteQueens,
		'K': &chess.WhiteKing,
		'p': &chess.BlackPawns,
		'r': &chess.BlackRooks,
		'n': &chess.BlackKnights,
		'b': &chess.BlackBishops,
		'q': &chess.BlackQueens,
		'k': &chess.BlackKing,
	}
	boardIndex := 63
	for _, piece := range data[0] {
		if '0' <= piece && piece <= '9' {
			//If meet a number, then we skip through squares
			boardIndex -= int(piece - '0')
		} else {
			if piece != '/' {
				*conv[piece] |= 0x1 << boardIndex
				boardIndex--
			}
		}
	}

	//Get the side to move
	chess.SideToMove = data[1] == "w"

	//Calculate castling privilege
	if data[2] == "-" {
		chess.CastlingPrivilege = 0
	} else {
		for _, cs := range data[2] {
			if cs == 'K' {
				chess.CastlingPrivilege |= 8
			} else if cs == 'Q' {
				chess.CastlingPrivilege |= 4
			} else if cs == 'k' {
				chess.CastlingPrivilege |= 2
			} else if cs == 'q' {
				chess.CastlingPrivilege |= 1
			}
		}
	}

	//Register en passant target
	if data[3] == "-" {
		chess.EnPassantTarget = -1
	} else {
		chess.EnPassantTarget = FromAlgebraicToIndex(data[3])
	}

	//Register half move clock
	var err error
	chess.Halfmove, err = strconv.Atoi(data[4])
	if err != nil {
		chess.Halfmove = 0
	}

	//Register full move counter
	chess.Fullmove, err = strconv.Atoi(data[5])
	if err != nil {
		chess.Fullmove = 1
	}
}

// Method to perform a deep copy on chess struct
func (chess *Chess) Clone() *Chess {
	clone := NewChess()

	clone.WhitePawns = chess.WhitePawns
	clone.WhiteRooks = chess.WhiteRooks
	clone.WhiteKnights = chess.WhiteKnights
	clone.WhiteBishops = chess.WhiteBishops
	clone.WhiteQueens = chess.WhiteQueens
	clone.WhiteKing = chess.WhiteKing

	clone.BlackPawns = chess.BlackPawns
	clone.BlackRooks = chess.BlackRooks
	clone.BlackKnights = chess.BlackKnights
	clone.BlackBishops = chess.BlackBishops
	clone.BlackQueens = chess.BlackQueens
	clone.BlackKing = chess.BlackKing

	clone.CastlingPrivilege = chess.CastlingPrivilege
	clone.EnPassantTarget = chess.EnPassantTarget
	clone.Fullmove = chess.Fullmove
	clone.Halfmove = chess.Halfmove
	clone.SideToMove = chess.SideToMove

	return clone
}

func (chess *Chess) Copy(c *Chess) {
	chess.WhitePawns = c.WhitePawns
	chess.WhiteRooks = c.WhiteRooks
	chess.WhiteKnights = c.WhiteKnights
	chess.WhiteBishops = c.WhiteBishops
	chess.WhiteQueens = c.WhiteQueens
	chess.WhiteKing = c.WhiteKing
	chess.BlackPawns = c.BlackPawns
	chess.BlackRooks = c.BlackRooks
	chess.BlackKnights = c.BlackKnights
	chess.BlackBishops = c.BlackBishops
	chess.BlackQueens = c.BlackQueens
	chess.BlackKing = c.BlackKing
	chess.CastlingPrivilege = c.CastlingPrivilege
	chess.EnPassantTarget = c.EnPassantTarget
	chess.Fullmove = c.Fullmove
	chess.Halfmove = c.Halfmove
	chess.SideToMove = c.SideToMove
}

// Method to return a string array representation of the chessboard
// Useful when it's not resource-intense operation like display
func (chess *Chess) ToArray() [64]string {
	board := [64]string{}
	var piece string

	//Hash map for bitboard to string
	conv := map[uint64]string{
		chess.WhitePawns:   "P",
		chess.WhiteRooks:   "R",
		chess.WhiteKnights: "N",
		chess.WhiteBishops: "B",
		chess.WhiteQueens:  "Q",
		chess.WhiteKing:    "K",
		chess.BlackPawns:   "p",
		chess.BlackRooks:   "r",
		chess.BlackKnights: "n",
		chess.BlackBishops: "b",
		chess.BlackQueens:  "q",
		chess.BlackKing:    "k",
	}

	//Loop through 64 squares in the board, and enter value to the string array
	for i := 0; i < 64; i++ {
		piece = " "
		for key, value := range conv {
			if IsPieceAtIndex(key, i) {
				piece = value
				break
			}
		}
		board[63-i] = piece
	}

	return board
}

// Method for printing the chess board
func (chess *Chess) Print() {
	board := chess.ToArray()

	//Convert to string
	boardStr := "+---+---+---+---+---+---+---+---+\n"
	for i := 0; i < 8; i++ {
		boardStr += fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s | %d\n",
			board[8*i+0], board[8*i+1], board[8*i+2], board[8*i+3], board[8*i+4], board[8*i+5], board[8*i+6], board[8*i+7], 8-i)
		boardStr += "+---+---+---+---+---+---+---+---+\n"
	}
	boardStr += "  A   B   C   D   E   F   G   H"
	fmt.Println(boardStr)
}
