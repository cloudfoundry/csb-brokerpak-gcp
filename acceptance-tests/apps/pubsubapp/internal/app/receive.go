package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
)

func handleReceive(w http.ResponseWriter, r *http.Request, client *pubsub.Client, subscriptionName string) {
	sub := client.Subscription(subscriptionName)

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(msg.Data)
		if err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}

		msg.Ack()

		log.Println("Receive done.")
	})
	if err != nil {
		fail(w, http.StatusInternalServerError, "sub.Receive: %v", err)
		return
	}
}
