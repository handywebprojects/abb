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

func StoreBookPosition(b Book, p BookPosition){
	bd := bookcoll.Doc(b.Id()).Collection("booklets").Doc(b.Bookletid(p.Fen))
	pc := bd.Collection("positions")
	bd.Set(ctx, map[string]interface{}{
		"positions": pc,
	})
	pc.Doc(p.Posid()).Set(ctx, p.Serialize())
}

func Synccache(b *Book){
	start := time.Now()
	fmt.Println(SEP)
	fmt.Println("syncing cache", b.Fullname())	
	fmt.Println(SEP)
	b.Poscache = make(map[string]BookPosition)
	numpos := 0
	iter := bookcoll.Doc(b.Id()).Collection("booklets").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		piter := doc.Ref.Collection("positions").Documents(ctx)
		fmt.Println("syncing", doc.Ref.ID, b.Fullname())
		for {
			pdoc, perr := piter.Next()
			if perr == iterator.Done {
				break
			}
			blob := pdoc.Data()["blob"].(string)
			p := BookPositionFromBlob(blob)
			b.Poscache[p.Posid()] = p
			numpos++
		}
	}
	elapsed := time.Since(start)
	fmt.Println("syncing cache done", b.Fullname(), "positions", numpos, "took", elapsed)
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