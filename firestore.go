////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import (
	"fmt"	
	"context"		

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
	bookcoll.Doc(b.Id()).Set(ctx, b.Serialize())
}

////////////////////////////////////////////////////////////////