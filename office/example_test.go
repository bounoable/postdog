package office_test

import (
	"context"
	"fmt"
	"log"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

func ExampleNew() {
	po := office.New(
		// Outgoing letter channel buffer size
		office.QueueBuffer(3),
		// Set the logger implementation
		office.WithLogger(office.NopLogger),
		// Add middleware
		office.WithMiddleware(
			office.MiddlewareFunc(func(
				ctx context.Context,
				let letter.Letter,
				next func(context.Context, letter.Letter) (letter.Letter, error),
			) (letter.Letter, error) {
				// Manipulate letter in some way
				let.Subject = fmt.Sprintf("Re: %s", let.Subject)
				return next(ctx, let)
			}),
		),
		// Add hooks
		office.WithSendHook(office.BeforeSend, func(_ context.Context, let letter.Letter) {
			fmt.Println(fmt.Sprintf("sending letter to %v: %s", let.To, let.Subject))
		}),
		// Add plugins
		office.WithPlugin(
			office.PluginFunc(func(ctx office.PluginContext) {
				ctx.Log("Installing plugin...")
				ctx.WithSendHook(office.AfterSend, func(_ context.Context, let letter.Letter) {
					fmt.Println(fmt.Sprintf("sent letter to %v: %s", let.To, let.Subject))
				})
			}),
		),
	)

	err := po.Send(context.Background(), letter.Write(
	// Letter options ...
	))

	if err != nil {
		log.Fatalf("could not send letter: %s", err)
	}
}
