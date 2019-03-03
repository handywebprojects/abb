////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import(
	"fmt"
	"strings"
	"strconv"
)

////////////////////////////////////////////////////////////////

const START_FEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

////////////////////////////////////////////////////////////////

func init(){
	fmt.Println("--> initializing board")
	fmt.Println("--> board initialized")
}

////////////////////////////////////////////////////////////////

type Piece struct{
	Kind string
	Color int
}

func (p Piece) Tostring() string{
	if p.Color == 0{
		return p.Kind
	}
	return strings.ToUpper(p.Kind)
}

type Board struct{
	Rep []Piece
	Turnfen string
	Castlefen string
	Epfen string
}

func (b Board) Tostring() string{
	rows := make([]string, 0)
	for i := 0; i < 8; i++{
		row := make([]string, 0)
		for j := 0; j < 8; j++{
			p := b.Rep[i*8+j]
			row = append(row, p.Tostring())
		}		
		rows = append(rows, strings.Join(row, " "))
	}
	posrep := strings.Join(rows, "\n")
	posrep += fmt.Sprintf("\n\n%s %s %s", b.Turnfen, b.Castlefen, b.Epfen)
	return posrep
}

func (b *Board) Setfromfen(fen string){
	b.Rep = make([]Piece, 64)
	fenparts := strings.Split(fen, " ")
	rawfen := fenparts[0]
	b.Turnfen = fenparts[1]
	b.Castlefen = fenparts[2]
	b.Epfen = fenparts[3]
	rawfenrows := strings.Split(rawfen, "/")
	cnt := 0
	for _, row := range(rawfenrows){
		cs := strings.Split(row, "")
		for _, c := range cs{
			if ( c >= "0" ) && ( c <= "9"){
				sc, _ := strconv.Atoi(c)
				for j := 0; j < sc; j++{
					b.Rep[cnt] = Piece{"-", 0}
					cnt++
				}
			}else{
				if ( c >= "A" ) && ( c <= "Z"){
					b.Rep[cnt] = Piece{strings.ToLower(c), 1}
					cnt++
				}else{
					b.Rep[cnt] = Piece{c, 0}
					cnt++
				}
			}
		}		
	}
}

func (b Board) Tofen() string{	
	buff := ""	
	scnt := 0
	for cnt :=0; cnt < 64; {
		p := b.Rep[cnt+scnt]				
		if p.Kind == "-"{
			scnt++
		}else{
			if(scnt>0){
				buff+=fmt.Sprintf("%d", scnt)
				cnt+=scnt
				scnt = 0
			}
			if p.Color == 1{
				buff+=strings.ToUpper(p.Kind)				
			}else{
				buff+=p.Kind
			}
			cnt++			
		}			
		if (scnt > 0) && (((cnt+scnt)%8) == 0){
			buff+=fmt.Sprintf("%d", scnt)
			cnt+=scnt
			scnt = 0
		}
		if(((cnt+scnt)%8)==0)&&(cnt<64){
			buff+="/"
		}
	}	
	buff += " " + b.Turnfen + " " + b.Castlefen + " " + b.Epfen + " 0 1"
	return buff
}

func Sqindeces(sq string) (int, int){	
	return int(sq[0]) - 97, int(56 - sq[1])
}

func index(i int, j int) int{
	return j*8 + i
}

func ijok(i int, j int) bool{
	if (i<0) || (i>7) || (j<0) || (j>7){
		return false
	}
	return true
}

func ijalgeb(i int, j int) string{
	return fmt.Sprintf("%c%c", 97+i, 56-j)
}

func (b *Board) Makealgebmove(algeb string){
	fromi, fromj := Sqindeces(algeb[0:2])
	toi, toj := Sqindeces(algeb[2:4])	
	fromindex := index(fromi, fromj)
	toindex := index(toi, toj)
	fromp := b.Rep[fromindex]	
	top := b.Rep[toindex]	
	if fromp.Kind == "p"{
		if ( fromj - toj ) == 2{
			if ijok(toi-1, toj){
				tp := b.Rep[index(toi-1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 0 ){
					b.Epfen = ijalgeb(toi, toj+1)
				}				
			}
			if ijok(toi+1, toj){
				tp := b.Rep[index(toi+1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 0 ){
					b.Epfen = ijalgeb(toi, toj+1)
				}				
			}
		}
		if ( toj - fromj ) == 2{
			if ijok(toi-1, toj){
				tp := b.Rep[index(toi-1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 1 ){
					b.Epfen = ijalgeb(toi, toj-1)
				}				
			}
			if ijok(toi+1, toj){
				tp := b.Rep[index(toi+1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 1 ){
					b.Epfen = ijalgeb(toi, toj-1)
				}				
			}
		}
	}
	b.Rep[fromindex] = Piece{"-", 0}
	b.Rep[toindex] = fromp
	cK := false
	cQ := false
	ck := false
	cq := false
	for i:=0;i<len(b.Castlefen);i++{
		cp := b.Castlefen[i:i+1]		
		if cp == "K"{
			cK = true
		}
		if cp == "Q"{
			cQ = true
		}
		if cp == "k"{
			ck = true
		}
		if cp == "q"{
			cq = true
		}
	}	
	if b.Turnfen == "w"{
		b.Turnfen = "b"
	}else{
		b.Turnfen = "w"
	}	
	if fromp.Kind == "k"{
		if algeb == "e1g1"{
			b.Rep[63] = Piece{"-", 0}
			b.Rep[61] = Piece{"r", 1}
			cK = false
			cQ = false
		}
		if algeb == "e1c1"{
			b.Rep[56] = Piece{"-", 0}
			b.Rep[59] = Piece{"r", 1}
			cK = false
			cQ = false
		}
		if algeb == "e8g8"{
			b.Rep[7] = Piece{"-", 0}
			b.Rep[5] = Piece{"r", 0}
			ck = false
			cq = false
		}
		if algeb == "e8c8"{
			b.Rep[0] = Piece{"-", 0}
			b.Rep[3] = Piece{"r", 0}
			ck = false
			cq = false
		}
	}		
	if len(algeb) == 5{
		b.Rep[toindex] = Piece{algeb[4:5], fromp.Color}
	}
	if top.Kind != "-"{
		b.Rep[toindex]=Piece{"-", 0}
		for di:=-1;di<2;di++{
			for dj:=-1;dj<2;dj++{
				if!((di==0)&&(dj==0)){
					ni := toi+di
					nj := toj+dj
					if ijok(ni, nj){
						cp := b.Rep[index(ni, nj)]
						if (cp.Kind != "-")&&(cp.Kind != "p"){
							b.Rep[index(ni, nj)] = Piece{"-", 0}
						}
					}
				}
			}
		}
	}
	if b.Rep[63].Kind == "-"{
		cK = false
	}
	if b.Rep[56].Kind == "-"{
		cQ = false
	}
	if b.Rep[7].Kind == "-"{
		ck = false
	}
	if b.Rep[0].Kind == "-"{
		cq = false
	}
	b.Castlefen = ""
	if cK{
		b.Castlefen+="K"
	}
	if cQ{
		b.Castlefen+="Q"
	}
	if ck{
		b.Castlefen+="k"
	}
	if cq{
		b.Castlefen+="q"
	}
	if b.Castlefen==""{
		b.Castlefen="-"
	}
}

////////////////////////////////////////////////////////////////
