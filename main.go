package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"serina/engine"
	"strings"
	"time"
)

var history []*engine.Chess

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

func main() {
	//Create chess instance
	chess := engine.NewChess()

	//Run forever until user choose to stop
	for {
		//Create reader to read from standard input
		reader := bufio.NewReader(os.Stdout)

		//Read the user command
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
			chess.Print()
		case "display":
			//Display the chessboard
			chess.Print()
		case "move_gen":
			//Generate all moves
			moves := chess.MoveGeneration()

			//Format to string and print to standard output
			str := "All moves available: ["
			for _, move := range moves {
				str += move + " ,"
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

			//Store the old value of chess for unmake
			history = append(history, chess.Clone()) //Sunce we store a pointer, we have to clone it

			//Make move and display
			chess.MakeMove(move)
			chess.Print()
		case "unmake":
			//Check if the history still have data for unmake
			if len(history) <= 0 {
				fmt.Println("No move for unmake")
			} else {
				//Copy the top most of the history
				chess.Copy(history[len(history)-1])
				//Remove the record
				history = history[:len(history)-1]
				//Print the board
				chess.Print()
			}
		case "perft":
			//Get the depth from user
			fmt.Print("Enter depth: ")
			var depth int
			fmt.Scanf("%d\n", &depth)

			//Perform perft
			start := time.Now()
			chess.FastPerft(depth)
			elapsed := time.Since(start)
			fmt.Printf("Took %d ms (%.2f seconds)\n", elapsed.Milliseconds(), elapsed.Seconds())
		case "clear":
			Clear()
		case "exit":
			return
		}
	}
}
