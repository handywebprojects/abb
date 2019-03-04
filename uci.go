////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import (	
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
	"math"
	"math/rand"
	"time"
)

////////////////////////////////////////////////////////////////

// constants for result filtering
const (
	HighestDepthOnly   uint = 1 << iota // only return the highest depth results
	IncludeUpperbounds uint = 1 << iota // include upperbound results
	IncludeLowerbounds uint = 1 << iota // include lowerbound results
)

// Options, for initializing the chess engine
type Options struct {
	UCI_Variant string // uci variant
	MultiPV int  // number of principal variations (ranks top X moves)
	Hash    int  // hash size in MB
	Ponder  bool // whether the engine should ponder
	OwnBook bool // whether the engine should use its opening book
	Threads int  // max number of threads the engine should use
}

// scoreKey helps us save the latest unique result where unique is
// defined as having unique values for each of the fields
type scoreKey struct {
	Depth      int
	MultiPV    int
	Upperbound bool
	Lowerbound bool
}

// ScoreResult holds the score result records returned
// by the engine
type ScoreResult struct {
	Time           int      // time spent to get this result (ms)
	Depth          int      // depth (number of plies) of result record
	SelDepth       int      // selective depth -- some engines don't report this
	Nodes          int      // total nodes searched to get this result
	NodesPerSecond int      // current nodes per second rate
	MultiPV        int      // 0 if MultiPV not set
	Lowerbound     bool     // true if reported as lowerbound
	Upperbound     bool     // true if reported as upperbound
	Score          int      // score centipawns or mate in X if Mate is true
	Mate           bool     // whether this move results in forced mate
	BestMoves      []string // best line for this result
}

// Results holds a slice of ScoreResult records
// as well as some overall result data
type Results struct {
	BestMove string
	results  map[scoreKey]ScoreResult
	Results  []ScoreResult
}

func (r Results) String() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return fmt.Sprintln(string(b))
}

// Engine holds the information needed to communicate with
// a chess engine executable. Engines should be created with
// a call to NewEngine(/path/to/executable)
type Engine struct {
	cmd    *exec.Cmd
	stdout *bufio.Reader
	stdin  *bufio.Writer
}

// NewEngine returns an Engine it has spun up
// and connected communication to
func NewEngine(path string) (*Engine, error) {
	eng := Engine{}
	eng.cmd = exec.Command(path)
	stdin, err := eng.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := eng.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := eng.cmd.Start(); err != nil {
		return nil, err
	}
	eng.stdin = bufio.NewWriter(stdin)
	eng.stdout = bufio.NewReader(stdout)
	return &eng, nil
}

// SetOptions sends setoption commands to the Engine
// for the values set in the Options record passed in
func (eng *Engine) SetOptions(opt Options) error {
	var err error
	err = eng.sendOption("UCI_Variant", opt.UCI_Variant)
	if err != nil {
		fmt.Println("could not send variant uci option", opt.UCI_Variant)
	}
	if opt.MultiPV > 0 {
		err = eng.sendOption("multipv", opt.MultiPV)
		if err != nil {
			return err
		}
	}
	if opt.Hash > 0 {
		err = eng.sendOption("hash", opt.Hash)
		if err != nil {
			return err
		}
	}
	if opt.Threads > 0 {
		err = eng.sendOption("threads", opt.Threads)
		if err != nil {
			return err
		}
	}
	err = eng.sendOption("ownbook", opt.OwnBook)
	if err != nil {
		return err
	}
	err = eng.sendOption("ponder", opt.Ponder)
	if err != nil {
		return err
	}
	return err
}

func (eng *Engine) sendOption(name string, value interface{}) error {
	_, err := eng.stdin.WriteString(fmt.Sprintf("setoption name %s value %v\n", name, value))
	if err != nil {
		return err
	}
	err = eng.stdin.Flush()
	return err
}

// SetFEN takes a FEN string and tells the engine to set the position
func (eng *Engine) SetFEN(fen string) error {
	_, err := eng.stdin.WriteString(fmt.Sprintf("position fen %s\n", fen))
	if err != nil {
		return err
	}
	err = eng.stdin.Flush()
	return err
}

// Go can use search moves, depth and time to move as filter  for the results being returned.
// see http://wbec-ridderkerk.nl/html/UCIProtocol.html
func (eng *Engine) Go(depth int, searchmoves string, movetime int, resultOpts ...uint) (*Results, error) {
	res := Results{}
	resultOpt := uint(0)
	if len(resultOpts) == 1 {
		resultOpt = resultOpts[0]
	}
	goCmd := "go "

	if depth != 0 {
		goCmd += fmt.Sprintf("depth %d", depth)
	}
	if searchmoves != "" {
		goCmd += fmt.Sprintf(" searchmoves %s", searchmoves)
	}
	if movetime != 0 {
		goCmd += fmt.Sprintf(" movetime %d", movetime)
	}
	goCmd += "\n"
	_, err := eng.stdin.WriteString(goCmd)

	if err != nil {
		return nil, err
	}
	err = eng.stdin.Flush()
	if err != nil {
		return nil, err
	}
	for {
		line, err := eng.stdout.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.Trim(line, "\n")
		if strings.HasPrefix(line, "bestmove") {
			dummy := ""
			_, err := fmt.Sscanf(line, "%s %s", &dummy, &res.BestMove)
			if err != nil {
				return nil, err
			}
			break
		}

		err = res.addLineToResults(line)
		if err != nil {
			return nil, err
		}
	}
	for _, v := range res.results {
		if resultOpt&HighestDepthOnly != 0 && v.Depth != depth {
			continue
		}
		if resultOpt&IncludeUpperbounds == 0 && v.Upperbound {
			continue
		}
		if resultOpt&IncludeLowerbounds == 0 && v.Lowerbound {
			continue
		}
		res.Results = append(res.Results, v)
	}
	sort.Sort(byDepth(res.Results))
	return &res, nil
}

// GoDepth takes a depth and an optional uint flag that configures filters
// for the results being returned.
func (eng *Engine) GoDepth(depth int, resultOpts ...uint) (*Results, error) {
	return eng.Go(depth, "", 0, resultOpts...)
}

type byDepth []ScoreResult

func (a byDepth) Len() int      { return len(a) }
func (a byDepth) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDepth) Less(i, j int) bool {
	if a[i].Depth == a[j].Depth {
		if a[i].MultiPV == a[j].MultiPV {
			if a[i].Lowerbound == a[j].Lowerbound {
				return a[i].Upperbound && !a[j].Upperbound
			}
			return a[i].Lowerbound && !a[j].Lowerbound
		}
		return a[i].MultiPV < a[j].MultiPV
	}
	return a[i].Depth < a[j].Depth
}

func (res *Results) addLineToResults(line string) error {
	var err error
	if !strings.HasPrefix(line, "info") {
		return nil
	}
	//log.Println(line)
	rd := strings.NewReader(line)
	s := scanner.Scanner{}
	s.Init(rd)
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanInts
	r := ScoreResult{}
	for s.Scan() != scanner.EOF {
		switch s.TokenText() {
		case "info":
		case "currmove":
			return nil
		case "depth":
			s.Scan()
			r.Depth, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "seldepth":
			s.Scan()
			r.SelDepth, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "time":
			s.Scan()
			r.Time, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "nodes":
			s.Scan()
			r.Nodes, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "nps":
			s.Scan()
			r.NodesPerSecond, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "multipv":
			s.Scan()
			r.MultiPV, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
		case "lowerbound":
			s.Scan()
			r.Lowerbound = true
		case "upperbound":
			s.Scan()
			r.Upperbound = true
		case "score":
			s.Scan()
			switch s.TokenText() {
			case "cp":
				s.Scan()
			case "mate":
				r.Mate = true
				s.Scan()
			}
			negative := 1
			if s.TokenText() == "-" {
				negative = -1
				s.Scan()
			}
			r.Score, err = strconv.Atoi(s.TokenText())
			if err != nil {
				return err
			}
			r.Score = r.Score * negative
		case "pv":
			for s.Scan() != scanner.EOF {
				r.BestMoves = append(r.BestMoves, s.TokenText())
			}
		}
	}
	if r.Depth > 0 {
		if res.results == nil {
			res.results = make(map[scoreKey]ScoreResult)
		}
		res.results[scoreKey{
			Depth:      r.Depth,
			MultiPV:    r.MultiPV,
			Upperbound: r.Upperbound,
			Lowerbound: r.Lowerbound,
		}] = r
	}
	return nil
}

func (eng *Engine) Close() {
	_, err := eng.stdin.WriteString("stop\n")
	if err != nil {
		log.Println("failed to stop engine:", err)
	}
	eng.stdin.Flush()
	err = eng.cmd.Process.Kill()
	if err != nil {
		log.Println("failed to kill engine:", err)
	}
	eng.cmd.Wait()
}

////////////////////////////////////////////////////////////////

var eng *Engine

////////////////////////////////////////////////////////////////

const INF_SCORE = 10000
const MATE_SCORE = 9000

////////////////////////////////////////////////////////////////

func init(){
	fmt.Println("--> initializing engine")
	eng, _ = NewEngine("engines/stockfish9")
	fmt.Println("--> engine initialized")
}

////////////////////////////////////////////////////////////////

func Analyze(fen string, depth int, variantkey string) BookPosition {
	eng.SetOptions(Options{
		UCI_Variant:variantkey,
		Hash:64,
		Threads:1,
		MultiPV:250,		
	})

	eng.SetFEN(fen)
	
	resultOpts := HighestDepthOnly
	results, _ := eng.GoDepth(depth, resultOpts)

	moves := results.Results
	p := NewPosition(fen)
	for _, move := range(moves){		
		score := move.Score
		depth := move.Depth
		p.Enginedepth = depth
		if move.Mate{
			if score < 0{
				score = -INF_SCORE - score
			}else{
				score = INF_SCORE - score
			}
		}else if math.Abs(float64(score)) > MATE_SCORE{
			if score < 0{
				score = -MATE_SCORE
			}else{
				score = MATE_SCORE
			}
		}		
		algeb := move.BestMoves[0]
		m := BookMove{algeb, score, score, INFINITE_MINIMAX_DEPTH, false}
		p.Moves[algeb] = m
	}

	return p
}

////////////////////////////////////////////////////////////////

func (b Book) SelectRecursive(fen string, depth int, line []string) string{
	fmt.Println("selecting fen", depth, line, fen)
	if depth > b.Analysisdepth{
		fmt.Println("max depth exceeded")
		return ""
	}
	p, ok := b.Getpos(fen)
	if ok{
		mli := p.Getmovelist().Items
		maxmoves := 1
		if depth < len(b.Widths){
			maxmoves = b.Widths[depth]
		}else{
			if len(b.Widths) > 0{
				maxmoves = b.Widths[len(b.Widths) - 1]
			}
		}
		// TODO: get maxmoves from book
		sel := rand.Intn(maxmoves)
		selmove := mli[sel]
		// cutoff
		if ( selmove.Score < -b.Cutoff ) || ( selmove.Score > b.Cutoff ){
			fmt.Println("cutoff")
			return ""
		}
		newfen := b.Makealgebmove(selmove.Algeb, fen)
		return b.SelectRecursive(newfen, depth + 1, append(line, selmove.Algeb))
	}else{
		fmt.Println("selected", fen)
		return fen
	}
}

func (b Book) Select() string{
	return b.SelectRecursive(b.Rootfen, 0, []string{})
}

func (b Book) Addone() string{
	fmt.Println("add one to", b.Fullname())
	fen := b.Select()
	if fen == "" {
		fmt.Println("add one failed")
		return ""
	}else{
		fmt.Println("analyzing", fen)
		p := Analyze(fen, int(b.Enginedepth), b.Variantkey)
		fmt.Println("storing", p.Posid())
		b.StorePosition(p)
		return fen
	}	
}

////////////////////////////////////////////////////////////////

func (b *Book) Minimaxrecursive(fen string, line []string, posids []string, depth int, maxdepth int, seldepth int, nodes int, cutoff int) (int, int, int){
	//fmt.Println("minimax", fen, line, posids, depth, maxdepth)
	max := -INF_SCORE
	// max depth exceeded
	if depth > maxdepth{		
		return 2 * max, seldepth, nodes
	}
	posid := Fen2posid(fen)
	// repetition
	for _, testposid := range posids{
		if testposid == posid{
			return 0, seldepth, nodes	
		}
	}
	newposids := append(posids, posid)
	// check if position is found
	p, ok := b.Poscache[posid]	
	if !ok{		
		return 2 * max, seldepth, nodes
	}	
	if depth > seldepth{
		seldepth = depth
	}
	nodes += 1
	for _, mi := range p.Getmovelist().Items{		
		// cutoff
		algeb := mi.Algeb
		value := mi.Score		
		haspv := false
		if ( mi.Score >= -cutoff ) && ( mi.Score <= cutoff ){
			newfen := b.Makealgebmove(algeb, fen)
			value, seldepth, nodes = b.Minimaxrecursive(newfen, append(line, algeb), newposids, depth + 1, maxdepth, seldepth, nodes, cutoff)			
			if value >= -INF_SCORE{
				haspv = true
			}
		}
		// failed node
		if value < -INF_SCORE{
			value = mi.Score
		}
		// don't overwrite eval of low depth nodes
		if depth < mi.Minimaxdepth{			
			p.Moves[algeb] = BookMove{algeb, mi.Score, value, depth, haspv}			
		}			
		if depth == 0{
			fmt.Println(algeb, mi.Score, value)	
		}		
		if value > max{
			max = value
		}
	}
	return -max, seldepth, nodes
}

func (b *Book) Minimaxout(){
	start := time.Now()
	fmt.Println(SEP)
	fmt.Println("minimaxing out", b.Fullname())	
	fmt.Println(SEP)
	value, seldepth, nodes := b.Minimaxrecursive(b.Rootfen, []string{}, []string{}, 0, b.Analysisdepth, 0, 0, b.Cutoff)
	elapsed := time.Since(start)
	fmt.Println("minimaxing done", b.Fullname(), -value, seldepth, nodes, "took", elapsed, "rate", float32(nodes) / float32(elapsed) * 1e9)	
}

////////////////////////////////////////////////////////////////