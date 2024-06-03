package app

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func handlePublish(w http.ResponseWriter, r *http.Request, client *pubsub.Client, topicName string) {
	ctx := context.Background()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "could not read body: %q", err)
		return
	}
	defer r.Body.Close()
	body := string(data)

	t := client.Topic(topicName)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(body),
	})

	var finished sync.WaitGroup
	finished.Add(1)

	go func(res *pubsub.PublishResult) {
		// The Get method blocks until a server-generated ID or
		// an error is returned for the published message.
		id, err := res.Get(ctx)
		if err != nil {
			// Error handling code can be added here.
			fmt.Fprintf(w, "Failed to publish: %v", err)
			return
		}
		fmt.Fprintf(w, "Published message msg ID: %v\n", id)
		finished.Done()
	}(result)

	w.WriteHeader(http.StatusCreated)
}
