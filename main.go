package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"serina/engine"
	"serina/web-ui/server"
	"strings"
	"time"
)

func Clear() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func CLI() {
	//Create chess instance
	chess := engine.NewChess()

	//Run forever until user choose to stop
	for {
		//Create reader to read from standard input
		reader := bufio.NewReader(os.Stdout)

		//Read the user command
		fmt.Print("Enter command: ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading from standard input\nError: %v\n", err)
			os.Exit(1)
		}
		cmd = strings.TrimSpace(cmd)

		//For each command, run the corresponding operation
		switch cmd {
		case "FEN":
			//Ask for FEN from user
			fmt.Print("Enter FEN: ")
			fen, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading from standard input\nError: %v\n", err)
				os.Exit(1)
			}
			fen = strings.TrimSpace(fen)

			//Import FEN and display the chessboard
			chess.FEN(fen)
			fmt.Println(chess)
		case "display":
			//Display the chessboard
			fmt.Println(chess)
		case "move_gen":
			//Generate all moves
			moves := chess.MoveGeneration()

			//Format to string and print to standard output
			fmt.Println("Number of moves: ", len(moves))
			str := "All moves available: ["
			for _, move := range moves {
				str += move.String() + ", "
			}
			str = str[:len(str)-1] + "]"
			fmt.Println(str)
		case "move":
			//Get move from user
			fmt.Print("Enter move: ")
			move, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading from standard input\nError: %v\n", err)
				os.Exit(1)
			}
			move = strings.TrimSpace(move)

			//Make move and display
			chess.MakeMove(engine.NewMove(chess, move))
			fmt.Println(chess)
		case "perft":
			//Get the depth from user
			fmt.Print("Enter depth: ")
			var depth int
			fmt.Scanf("%d\n", &depth)

			//Perform perft
			start := time.Now()
			res, total := chess.FastPerft(depth)
			elapsed := time.Since(start)
			for key, val := range res {
				fmt.Printf("%s: %d\n", key, val)
			}
			fmt.Printf("Total node found: %d\n", total)
			fmt.Printf("Took %d ms (%.2f seconds)\n", elapsed.Milliseconds(), elapsed.Seconds())
		case "evaluate":
			fmt.Println("Current position evaluation: ", chess.Evaluate())
		case "search":
			//Get the depth from user
			fmt.Print("Enter depth: ")
			var depth int
			fmt.Scanf("%d\n", &depth)

			//Perform search
			start := time.Now()
			_, searchedMove := chess.Search(depth, -math.MaxInt32, math.MaxInt32)
			elapsed := time.Since(start)
			fmt.Println("Found move: ", searchedMove)
			fmt.Printf("Took %d ms (%.2f seconds)\n", elapsed.Milliseconds(), elapsed.Seconds())
		case "test":
			//Get the depth from user
			fmt.Print("Enter depth: ")
			var depth int
			fmt.Scanf("%d\n", &depth)

			//Perform search
			start := time.Now()
			total := chess.Perft(depth)
			elapsed := time.Since(start)
			fmt.Println("Total nodes: ", total)
			fmt.Printf("Took %d ms (%.2f seconds)\n", elapsed.Milliseconds(), elapsed.Seconds())
		case "clear":
			Clear()
		case "exit":
			return
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		CLI()
	} else {
		server := server.NewServer()
		server.Start()
	}
}
