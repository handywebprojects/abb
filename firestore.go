////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import (
	"fmt"	
	"context"		
	"time"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"cloud.google.com/go/firestore"
)

////////////////////////////////////////////////////////////////

var ctx context.Context
var client *firestore.Client
var bookcoll *firestore.CollectionRef

////////////////////////////////////////////////////////////////

const BOOK_ROOT = "books2"

////////////////////////////////////////////////////////////////

func init(){
	fmt.Println("--> initializing firestore")
	ctx = context.Background()
	opt := option.WithCredentialsFile("firebase/fbsacckey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil{
		fmt.Println("Fatal. Firestore app could not be initialized.")
		return
	}
	client, err = app.Firestore(ctx)
	if err != nil{
		fmt.Println("Fatal. Firestore client could not be created.")
		return
	}
	bookcoll = client.Collection(BOOK_ROOT)
	testbook := bookcoll.Doc("test")
	value, err := testbook.Set(ctx, map[string]interface{}{
		"meta": "test",
	})
	fmt.Println("testbook", value, err)
	fmt.Println("--> firestore initialized")
}

func Listbooks(){
	fmt.Printf("list of books [ root : %s ]\n", BOOK_ROOT)
	iter := bookcoll.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		fmt.Println("*",doc.Ref.ID)
	}
}

func StoreBook(b Book){	
	bd := bookcoll.Doc(b.Id())
	b.Booklets = bd.Collection("booklets")
	bd.Set(ctx, b.Serialize())
}

func Synccache(b *Book){
	start := time.Now()
	fmt.Println(SEP)
	fmt.Println("syncing cache", b.Fullname())	
	fmt.Println(SEP)
	b.Poscache = make(map[string]BookPosition)
	numpos := 0
	grandtotalblobsize := 0
	maxnumbpos := 0
	maxtotalblobsize := 0
	iter := bookcoll.Doc(b.Id()).Collection("booklets").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}		
		positions := doc.Data()["positions"].(map[string]interface{})
		numbpos := 0
		totalblobsize := 0
		for _, posdoc := range(positions){			
			blob := posdoc.(map[string]interface{})["blob"].(string)
			p := BookPositionFromBlob(blob)
			b.Poscache[p.Posid()] = p
			numpos++
			numbpos++
			totalblobsize += len(blob)
			grandtotalblobsize += len(blob)
		}		
		fmt.Println(doc.Ref.ID, numbpos, totalblobsize)
		if numbpos > maxnumbpos{
			maxnumbpos = numbpos			
		}
		if totalblobsize > maxtotalblobsize{
			maxtotalblobsize = totalblobsize
		}
	}
	elapsed := time.Since(start)
	fmt.Println("syncing cache done", b.Fullname(), "positions", numpos, "took", elapsed, "average blob size", grandtotalblobsize / (numpos+1), "max positions per booklet", maxnumbpos, "max total blobsize", maxtotalblobsize)	
	bookcoll.Doc(b.Id()).Update(ctx, []firestore.Update{
		{Path: "numpos", Value: numpos},
		{Path: "maxnumbpos", Value: maxnumbpos},
		{Path: "maxtotalblobsize", Value: maxtotalblobsize},
	})
}

func Uploadcache(b Book){
	start := time.Now()
	fmt.Println(SEP)
	fmt.Println("uploading cache", b.Fullname())	
	fmt.Println(SEP)
	numpos := 0
	booklets := make(map[string]map[string]interface{})
	for _, p := range(b.Poscache){		
		bid := b.Bookletid(p.Fen)
		booklet, ok := booklets[bid]
		if !ok{
			booklet = map[string]interface{}{
				"positions": make(map[string]interface{}),
			}
			booklets[bid] = booklet
		}
		booklet, _ = booklets[bid]		
		booklet["positions"].(map[string]interface{})[p.Posid()] = p.Serialize()
		numpos++
	}		
	for bookletid, booklet := range(booklets){
		fmt.Println("uploading", bookletid, b.Fullname())		
		bookcoll.Doc(b.Id()).Collection("booklets").Doc(bookletid).Set(ctx, booklet)
	}
	elapsed := time.Since(start)
	fmt.Println("uploading cache done", b.Fullname(), "positions", numpos, "took", elapsed)
}

////////////////////////////////////////////////////////////////