package main

import (
    "fmt"
)

func main() {
    fmt.Printf("Game started\n")

    game := NewGame()
    game.GameLoop()

    fmt.Printf("Exiting\n")
}
