package main

import (
	"fmt"

	"github.com/rashedmyt/go-turtlecoin/crypto/keccak"
)

func main() {
	fmt.Println("Its a test")
	seed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		seed[i] = byte(i)
	}
	response := keccak.Keccak(seed, 32)
	fmt.Println(response)
}
