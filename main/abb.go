package main

import(
	"fmt"	

	"github.com/handywebprojects/abb"
)

func main(){
	fmt.Println("abb - Auto Book Builder")	
	//fmt.Println(abb.NewBoard("atomic").Tostring())
	b := abb.NewBook()
	b.Store()
	abb.Listbooks()
}
