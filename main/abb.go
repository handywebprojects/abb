package main

import(
	"fmt"	
	"time"

	"github.com/handywebprojects/abb"
)

func main(){		
	fmt.Println("abb - Auto Book Builder")		
	b := abb.NewBook()		
	b.Store()	
	abb.Listbooks()
	b.Synccache()	
	b.Minimaxout()
	b.Uploadcache()
	//return
	time.Sleep(3 * time.Second)
	for i:=0; i<b.Numcycles; i++{
		fmt.Println(abb.SEP)
		fmt.Println("build cycle", i+1, "of", b.Numcycles)
		fmt.Println(abb.SEP)
		time.Sleep(3 * time.Second)
		for j:=0; j<b.Batchsize; j++{
			fmt.Println(abb.SEP)
			fmt.Println("batch", j+1, "of", b.Batchsize, "of build cycle", i+1, "of", b.Numcycles)
			fmt.Println(abb.SEP)
			time.Sleep(1 * time.Second)
			b.Addone()
			if ( j % b.Minimaxafter ) == 0{
				b.Minimaxout()
			}
		}
		b.Minimaxout()
		b.Uploadcache()
	}
}
