package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Chess struct
type Chess struct {
	Boards            [12]uint64
	SideToMove        int
	EnPassantTarget   int // Range from 0 to 63
	CastlingPrivilege int // Use 4 bit integer to represent: KQkq (exactly in this order)
	Halfmove          int
	Fullmove          int
}

func NewChess() *Chess {
	return &Chess{
		Boards:            [12]uint64{}, //Empty boards
		SideToMove:        WHITE,        //White turn as default
		EnPassantTarget:   -1,           //No en passant target
		CastlingPrivilege: 0,            //No castling privilege
		Halfmove:          0,            //Default halfmove
		Fullmove:          1,            //Default fullmove
	}
}

func (chess *Chess) Clear() {
	for i := WHITE_PAWN; i <= BLACK_KING; i++ {
		chess.Boards[i] = 0
	}
	chess.EnPassantTarget = -1
	chess.CastlingPrivilege = 0
	chess.SideToMove = WHITE
	chess.Halfmove = 0
	chess.Fullmove = 1
}

func (chess *Chess) FEN(fen string) {
	//Clear the old data of chess object (in case of replay)
	chess.Clear()

	//If empty string is provided, then we assumed it to be default position
	if len(fen) == 0 || fen == "" {
		fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" //Starting position
	}

	//Split the fen string
	data := strings.Split(fen, " ")

	//Set up the bitboards
	pieceMapping := map[rune]int{
		'P': WHITE_PAWN,
		'R': WHITE_ROOK,
		'N': WHITE_KNIGHT,
		'B': WHITE_BISHOP,
		'Q': WHITE_QUEEN,
		'K': WHITE_KING,
		'p': BLACK_PAWN,
		'r': BLACK_ROOK,
		'n': BLACK_KNIGHT,
		'b': BLACK_BISHOP,
		'q': BLACK_QUEEN,
		'k': BLACK_KING,
	}
	boardIndex := 63
	for _, piece := range data[0] {
		if '1' <= piece && piece <= '8' { //A valid syntax FEN string can only have digit from 1 to 8
			boardIndex -= int(piece - '0')
		} else {
			if piece != '/' {
				chess.Boards[pieceMapping[piece]] |= 0x1 << boardIndex
				boardIndex--
			}
		}
	}

	//Get the side to move
	if data[1] == "w" {
		chess.SideToMove = WHITE
	} else {
		chess.SideToMove = BLACK
	}

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

	//If the FEN string didn't provide the halfmove and fullmove value, we'll assign fallback value to them
	if len(data) <= 4 {
		chess.Halfmove = 0
		chess.Fullmove = 1
		return
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

func (chess *Chess) Clone() *Chess {
	clone := NewChess()

	for i := WHITE_PAWN; i <= BLACK_KING; i++ {
		clone.Boards[i] = chess.Boards[i]
	}
	clone.CastlingPrivilege = chess.CastlingPrivilege
	clone.EnPassantTarget = chess.EnPassantTarget
	clone.Fullmove = chess.Fullmove
	clone.Halfmove = chess.Halfmove
	clone.SideToMove = chess.SideToMove

	return clone
}

func (chess *Chess) Copy(c *Chess) {
	for i := WHITE_PAWN; i <= BLACK_KING; i++ {
		chess.Boards[i] = c.Boards[i]
	}
	chess.CastlingPrivilege = c.CastlingPrivilege
	chess.EnPassantTarget = c.EnPassantTarget
	chess.Fullmove = c.Fullmove
	chess.Halfmove = c.Halfmove
	chess.SideToMove = c.SideToMove
}

func (chess *Chess) Flip() {
	for i := range 6 {
		chess.Boards[i], chess.Boards[i+6] = FlipVertical(chess.Boards[i+6]), FlipVertical(chess.Boards[i])
	}

	//Flip side to move
	if chess.SideToMove == WHITE {
		chess.SideToMove = BLACK
	} else {
		chess.SideToMove = WHITE
	}

	//Flip the en passant target (if any)
	if chess.EnPassantTarget != -1 {
		chess.EnPassantTarget = FlipIndexVertical(chess.EnPassantTarget)
	}

	//Flip castling privilege
	chess.CastlingPrivilege = ((chess.CastlingPrivilege >> 2) | (chess.CastlingPrivilege << 2)) & 15
}

func (chess *Chess) ToArray() [64]string {
	board := [64]string{}
	var piece string

	pieceMapping := [12]string{
		"P", "R", "N", "B", "Q", "K", "p", "r", "n", "b", "q", "k",
	}

	//Loop through 64 squares in the board, and enter value to the string array
	for index := range 64 {
		piece = " "
		for i := WHITE_PAWN; i <= BLACK_KING; i++ {
			if IsPieceAtIndex(chess.Boards[i], index) {
				piece = pieceMapping[i]
				break
			}
		}
		board[63-index] = piece
	}

	return board
}

func (chess *Chess) String() string {
	board := chess.ToArray()

	//Convert to string
	boardStr := "+---+---+---+---+---+---+---+---+\n"
	for i := range 8 {
		for j := range 8 {
			boardStr += fmt.Sprintf("| %s ", board[8*i+j])
		}
		boardStr += fmt.Sprintf("| %d\n", 8-i)
		boardStr += "+---+---+---+---+---+---+---+---+\n"
	}
	boardStr += "  A   B   C   D   E   F   G   H"

	boardStr += fmt.Sprintf("\nSide to move: %d\nEn passant target: %s\nCastling: %d\n",
		chess.SideToMove, FromIndexToAlgebraic(chess.EnPassantTarget), chess.CastlingPrivilege)

	return boardStr
}
