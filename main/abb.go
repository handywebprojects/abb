package main

import(
	"fmt"	

	"github.com/handywebprojects/abb"
)

func main(){		
	fmt.Println("abb - Auto Book Builder")		
	b := abb.NewBook()	
	b.Enginedepth = 5
	b.Store()	
	abb.Listbooks()
	b.Synccache()
	b.Addone()
}
