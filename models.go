////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import(
	"fmt"
	"strconv"
	"strings"
	"sort"

	"cloud.google.com/go/firestore"
)

////////////////////////////////////////////////////////////////

const INFINITE_MINIMAX_DEPTH = 1000

////////////////////////////////////////////////////////////////

type BookMove struct{
	Algeb string
	Score int
	Eval int
	Minimaxdepth int
	Haspv int
}

func (m BookMove) Serialize() string{
	return fmt.Sprintf("%s;%d;%d;%d;%d", m.Algeb, m.Score, m.Eval, m.Minimaxdepth, m.Haspv)
}

func BookMoveFromBlob(blob string) BookMove{
	parts := strings.Split(blob, ";")
	return BookMove{
		Algeb: parts[0],
		Score: str2int(parts[1], 0),
		Eval: str2int(parts[2], 0),
		Minimaxdepth: str2int(parts[3], INFINITE_MINIMAX_DEPTH),
		Haspv: str2int(parts[1], 0),
	}
}

type Movelist struct{
	Items []BookMove
}

func (m *Movelist) Len() int{
	return len(m.Items)
}

func (m *Movelist) Swap(i, j int){
	m.Items[i], m.Items[j] = m.Items[j], m.Items[i]
}

func (m *Movelist) Less(i, j int) bool{
	return m.Items[i].Eval > m.Items[j].Eval
}

type BookPosition struct{
	Fen string
	Enginedepth int
	Moves map[string]BookMove
}

func (p BookPosition) Posid() string{
	return Fen2posid(p.Fen)
}

func (p BookPosition) Getmovelist() Movelist{
	movelist := make([]BookMove, 0)
	for _, move := range(p.Moves){
		move.Minimaxdepth = INFINITE_MINIMAX_DEPTH
		movelist = append(movelist, move)
	}
	ml := Movelist{movelist}
	sort.Sort(&ml)
	return ml
}

func (p BookPosition) Serialize() map[string]interface{}{
	strs := []string{}
	for _, m := range(p.Moves){
		strs = append(strs, m.Serialize())
	}
	blob := fmt.Sprintf("%s;;%d;;%s", p.Fen, p.Enginedepth, strings.Join(strs, "|"))
	return map[string]interface{}{
		"blob": blob,
	}
}

func NewPosition(fen string) BookPosition{
	return BookPosition{
		Fen: fen,
		Moves: make(map[string]BookMove),
	}
}

func BookPositionFromBlob(blob string) BookPosition{
	parts := strings.Split(blob, ";;")
	moveparts := strings.Split(parts[2], "|")
	p := BookPosition{
		Fen: parts[0],
		Enginedepth: str2int(parts[1], 0),
		Moves: make(map[string]BookMove),
	}
	for _, moveblob := range(moveparts){
		m := BookMoveFromBlob(moveblob)
		p.Moves[m.Algeb] = m
	}
	return p
}

type Book struct{
	Name string
	Variantkey string
	Rootfen string
	Mod int
	Analysisdepth int
	Enginedepth int
	Numcycles int
	Batchsize int
	Minimaxafter int
	Cutoff int
	Widths []int	
	Booklets *firestore.CollectionRef
	Poscache map[string]BookPosition
}

func (b *Book) Synccache(){
	Synccache(b)
}

func (b Book) Uploadcache(){
	Uploadcache(b)
}

func (b Book) Id() string{
	return fmt.Sprintf("%s%s", b.Name, b.Variantkey)
}

func (b Book) Fullname() string{
	return fmt.Sprintf("[Book %s %s]", b.Name, b.Variantkey)
}

func NewBook() Book{
	return Book{
		Name: Envstr("BOOKNAME", "default"),
		Variantkey: Envstr("BOOKVARIANT", "atomic"),
		Rootfen: Envstr("ANALYSISROOT", START_FEN),
		Mod: Envint("BOOKMOD", 10),
		Analysisdepth: Envint("ANALYSISDEPTH", 20),
		Enginedepth: Envint("ENGINEDEPTH", 20),
		Numcycles: Envint("NUMCYCLES", 10),
		Batchsize: Envint("BATCHSIZE", 10),
		Minimaxafter: Envint("MINIMAXAFTER", 3),
		Cutoff: Envint("CUTOFF", 1000),
		Widths: Envintarray("WIDTHS", []int{3,2,1}),
		Poscache: make(map[string]BookPosition),
	}
}

func (b Book) Getpos(fen string) (BookPosition, bool){
	posid := Fen2posid(fen)
	p, ok := b.Poscache[posid]
	return p, ok
}

func (b Book) Serialize() map[string]interface{}{
	return map[string]interface{}{
		"name": b.Name,
		"variantkey": b.Variantkey,
		"rootfen": b.Rootfen,
		"mod": strconv.Itoa(b.Mod),
		"analysisdepth": strconv.Itoa(b.Analysisdepth),
		"enginedepth": strconv.Itoa(b.Enginedepth),
		"numcycles": strconv.Itoa(b.Numcycles),
		"batchsize": strconv.Itoa(b.Batchsize),
		"minimaxafter": strconv.Itoa(b.Minimaxafter),
		"cutoff": strconv.Itoa(b.Cutoff),
		"widths": Intarray2str(b.Widths),
		"booklets": b.Booklets,
	}
}

func (b Book) Store(){
	StoreBook(b)
}

func (b Book) Bookletid(fen string) string{
	return Bookletid(fen, b.Mod)
}

func (b Book) Analyze(fen string) BookPosition{
	return Analyze(fen, b.Enginedepth, b.Variantkey)
}

func (b Book) StorePosition(p BookPosition){
	b.Poscache[p.Posid()] = p	
}

////////////////////////////////////////////////////////////////