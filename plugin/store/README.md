# postdog - Store plugin

This plugin stores sent letters in a database and provides functions to query those letters.

## Example

```go
package main

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memorystore"
)

func main() {
  po := postdog.New(
		postdog.WithPlugin(
			store.Plugin(
				memorystore.New(), // in-memory store
			),
		),
	)

	err := po.Send(context.Background(), letter.Write(
		letter.Text("Hello."),
	))
}
```

## Disable insertion

You can disable the insertion of letters into the database through the context:

```go
package main

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memorystore"
)

func main() {
  po := postdog.New(
		postdog.WithPlugin(
			store.Plugin(
				memorystore.New(), // in-memory store
			),
		),
	)

	// disable storage for this context
  ctx := store.Disable(context.Background())

	err := po.Send(ctx, letter.Write(
		letter.Text("Hello."),
	))
}
```

## Query letters

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memorystore"
	"github.com/bounoable/postdog/plugin/store/query"
)

func main() {
  memstore := memorystore.New( /* fill the store with letters */ ) // in-memory store

	_ = postdog.New(
		postdog.WithPlugin(store.Plugin(memstore)),
	)

	ctx := context.Background()

	cur, err := query.Run(
		ctx,
		memstore,
		query.Subject("order", "offer"), // subject must contain "order" or "offer"
		query.SentBetween(time.Now().AddDate(0, 0, -7), time.Now()), // letter must have been sent in the past 7 days
		query.Sort(query.SortBySendDate, query.SortDesc),            // sort descending by send date
		// see the `query` package for more query options
	)
	if err != nil {
		panic(err)
	}
	defer cur.Close(ctx) // close the cursor after use

	for cur.Next(ctx) {
		let := cur.Current()
		fmt.Println(let)
	}

	if cur.Err() != nil {
		panic(cur.Err())
	}
}
```