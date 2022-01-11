package app

import (
	"cloud.google.com/go/spanner"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"spannerapp/internal/credentials"
)

func App(creds credentials.SpannerCredentials) *mux.Router {
	client, err := spanner.NewClient(context.Background(), creds.FullDBName, option.WithCredentialsJSON([]byte(creds.Credentials)))
	if err != nil {
		log.Printf(" Error creating client: %s", err)
	}

	_, err = client.ReadWriteTransaction(context.Background(), func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `CREATE TABLE IF NOT EXISTS test (
				keyname  STRING(1024),
				valuedata   STRING(1024),
			)`,
		}
		_, err := txn.Update(ctx, stmt)
		if err != nil {
			log.Printf("Error creating table - 1 : %s", err)
			return err
		}
		log.Printf("Error creating table - 2 : %s", err)
		return err
	})
	log.Printf("Error creating table: %s", err)

	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	//r.HandleFunc("/{fileName}", handleUpload(client, creds.BucketName)).Methods("PUT")
	//r.HandleFunc("/{fileName}", handleDownload(client, creds.BucketName)).Methods("GET")

	return r
}

//adminClient, err := database.NewDatabaseAdminClient(ctx, option.WithCredentialsJSON([]byte(creds.Credentials)))
//if err != nil {
//log.Printf("Error creating adminClient: %s", err)
//}
//defer adminClient.Close()
//
//op, err := adminClient.UpdateDatabaseDdl(ctx, &adminpb.UpdateDatabaseDdlRequest{
//Database: creds.DBName,
//Statements: []string{
//`CREATE TABLE IF NOT EXISTS test (
//				keyname  STRING(1024),
//				valuedata   STRING(1024),
//			)`,
//},
//})
//if err != nil {
//log.Printf("Error creating adminClient: %s", err)
//}
//if err := op.Wait(ctx); err != nil {
//log.Printf("Error creating adminClient: %s", err)
//}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
