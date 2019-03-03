package main

import(
	"fmt"	

	"github.com/handywebprojects/abb"
)

func main(){
	eng, err := abb.NewEngine("engines/stockfish9")
	fmt.Println(eng, err)
}
