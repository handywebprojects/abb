package main

import(
	"fmt"	

	"github.com/handywebprojects/abb"
)

func main(){
	b := abb.Board{}
	b.Setfromfen(abb.START_FEN)
	fmt.Println(b.Tostring())
	b.Makealgebmove("f7f4")
	b.Makealgebmove("e2e4")
	fmt.Println(b.Tostring())
}
