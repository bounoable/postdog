package query_test

import (
	"context"
	"fmt"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memory"
	"github.com/bounoable/postdog/plugin/store/query"
)

func ExampleRun() {
	// Usually this would be a persistent implementation
	repo := memory.NewStore(
		store.Letter{Letter: letter.Write(letter.Subject("Letter 1"))},
		store.Letter{Letter: letter.Write(letter.Subject("Letter 2"))},
	)

	cur, err := query.Run(
		context.Background(),
		repo,
		query.Subject("Letter"), // sender name / address must contain "Letter"
		query.Sort(query.SortBySendDate, query.SortDesc), // sort descending by send date
		// more query options ...
	)
	defer cur.Close(context.Background())

	if err != nil {
		panic(err)
	}

	for cur.Next(context.Background()) {
		let := cur.Current()
		fmt.Println(let)
	}

	if cur.Err() != nil {
		panic(cur.Err())
	}
}
