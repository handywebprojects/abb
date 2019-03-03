////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import(
	"fmt"
	"strconv"
)

////////////////////////////////////////////////////////////////

type Book struct{
	Name string
	Variantkey string
	Rootfen string
	Analysisdepth int
	Enginedepth int
	Numcycles int
	Batchsize int
	Cutoff int
	Widths []int	
}

func (b Book) Id() string{
	return fmt.Sprintf("%s%s", b.Name, b.Variantkey)
}

func NewBook() Book{
	return Book{
		Name: Envstr("BOOKNAME", "default"),
		Variantkey: Envstr("BOOKVARIANT", "atomic"),
		Rootfen: Envstr("ANALYSISROOT", START_FEN),
		Analysisdepth: Envint("ANALYSISDEPTH", 20),
		Enginedepth: Envint("ENGINEDEPTH", 20),
		Numcycles: Envint("NUMCYCLES", 10),
		Batchsize: Envint("BATCHSIZE", 10),
		Cutoff: Envint("CUTOFF", 500),
		Widths: Envintarray("WIDTHS", []int{3,2,1}),
	}
}

func (b Book) Serialize() map[string]interface{}{
	return map[string]interface{}{
		"name": b.Name,
		"variantkey": b.Variantkey,
		"rootfen": b.Rootfen,
		"analysisdepth": strconv.Itoa(b.Analysisdepth),
		"enginedepth": strconv.Itoa(b.Enginedepth),
		"numcycles": strconv.Itoa(b.Numcycles),
		"batchsize": strconv.Itoa(b.Batchsize),
		"cutoff": strconv.Itoa(b.Cutoff),
		"widths": Intarray2str(b.Widths),
	}
}

func (b Book) Store(){
	StoreBook(b)
}

////////////////////////////////////////////////////////////////