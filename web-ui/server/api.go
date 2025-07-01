package server

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"serina/engine"
	"strconv"
	"time"
)

type ChessData struct {
	Board             [64]string `json:"board"`
	SideToMove        string     `json:"side_to_move"`
	EnPassantTarget   string     `json:"en_passant_target"`
	CastlingPrivilege string     `json:"castling"`
	Halfmove          int        `json:"halfmove"`
	Fullmove          int        `json:"fullmove"`
	Moves             []string   `json:"moves"`
}

func NewChessData(chess *engine.Chess) *ChessData {
	var chessData *ChessData = &ChessData{}

	chessData.Board = chess.ToArray()
	if chess.SideToMove == engine.WHITE {
		chessData.SideToMove = "white"
	} else {
		chessData.SideToMove = "black"
	}
	chessData.EnPassantTarget = engine.FromIndexToAlgebraic(chess.EnPassantTarget)

	//Handle castling privilege conversion
	switch {
	case chess.CastlingPrivilege&int(engine.WHITE_KING_SIDE) == int(engine.WHITE_KING_SIDE):
		chessData.CastlingPrivilege += "K"
	case chess.CastlingPrivilege&int(engine.WHITE_QUEEN_SIDE) == int(engine.WHITE_QUEEN_SIDE):
		chessData.CastlingPrivilege += "Q"
	case chess.CastlingPrivilege&int(engine.WHITE_KING_SIDE) == int(engine.WHITE_KING_SIDE):
		chessData.CastlingPrivilege += "k"
	case chess.CastlingPrivilege&int(engine.WHITE_KING_SIDE) == int(engine.WHITE_KING_SIDE):
		chessData.CastlingPrivilege += "q"
	}

	chessData.Halfmove = chess.Halfmove
	chessData.Fullmove = chess.Fullmove
	moves := chess.MoveGeneration()
	for _, move := range moves {
		chessData.Moves = append(chessData.Moves, move.String())
	}
	// chessData.Moves = chess.MoveGeneration()

	return chessData
}

func (server *Server) HandleFEN(w http.ResponseWriter, r *http.Request) {
	//Get the FEN string from request
	params := r.URL.Query()
	fen := params.Get("fen")

	//Setup the chessboard using the FEN string and get its array representation
	server.chess.FEN(fen)

	//Send the data back as JSON
	data, err := json.MarshalIndent(NewChessData(server.chess), "", "")
	if err != nil {
		fmt.Printf("Error marshaling chessboard array to JSON\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (server *Server) HandleFlipBoard(w http.ResponseWriter, r *http.Request) {
	server.chess.Flip()

	//Send the data back as JSON
	data, err := json.MarshalIndent(NewChessData(server.chess), "", "")
	if err != nil {
		fmt.Printf("Error marshaling chessboard array to JSON\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (server *Server) HandleMove(w http.ResponseWriter, r *http.Request) {
	//Get the move from URL
	params := r.URL.Query()
	move := params.Get("move")

	//Perform the move
	server.chess.MakeMove(engine.NewMove(server.chess, move))

	//Send the data back as JSON
	data, err := json.MarshalIndent(NewChessData(server.chess), "", "")
	if err != nil {
		fmt.Printf("Error marshaling chessboard array to JSON\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

type PerftResult struct {
	Result    map[string]int `json:"result"`
	TotalNode int            `json:"total_node"`
	Time      int            `json:"time"`
}

func (server *Server) HandlePerft(w http.ResponseWriter, r *http.Request) {
	//Get the depth from URL
	params := r.URL.Query()
	if params.Get("depth") == "" {
		http.Error(w, "Missing request parameter 'depth'", http.StatusBadRequest)
		return
	}
	depth, err := strconv.Atoi(params.Get("depth"))
	if err != nil {
		fmt.Printf("Error parsing depth from request parameters\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//Get the perft result
	start := time.Now()
	result, totalNode := server.chess.FastPerft(depth)
	elapsed := time.Since(start)

	//Send the data back as JSON
	data := PerftResult{
		Result:    result,
		TotalNode: totalNode,
		Time:      int(elapsed.Milliseconds()),
	}

	jsonData, err := json.MarshalIndent(data, "", "")
	if err != nil {
		fmt.Printf("Error marshaling perft result to JSON\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

type SearchResult struct {
	SearchedMove string `json:"searched_move"`
	Time         int    `json:"time"`
}

func (server *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	//Get the depth from URL
	params := r.URL.Query()
	if params.Get("depth") == "" {
		http.Error(w, "Missing request parameter 'depth'", http.StatusBadRequest)
		return
	}
	depth, err := strconv.Atoi(params.Get("depth"))
	if err != nil {
		fmt.Printf("Error parsing depth from request parameters\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//Get the search result
	start := time.Now()
	_, searchedMove := server.chess.Search(depth, -math.MaxInt32, math.MaxInt32)
	elapsed := time.Since(start)

	//Send the data back as JSON
	data := SearchResult{
		SearchedMove: searchedMove.String(),
		Time:         int(elapsed.Milliseconds()),
	}

	jsonData, err := json.MarshalIndent(data, "", "")
	if err != nil {
		fmt.Printf("Error marshaling search result to JSON\nError: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}
