package app

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"pubsubapp/internal/credentials"
)

func App(creds credentials.PubSubCredentials) http.HandlerFunc {
	client, _ := pubsub.NewClient(context.Background(), creds.ProjectID, option.WithCredentialsJSON([]byte(creds.Credentials)))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleReceive(w, r, client, creds.SubscriptionName)
		case http.MethodPut:
			handlePublish(w, r, client, creds.TopicName)
		default:
			fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
	}
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
