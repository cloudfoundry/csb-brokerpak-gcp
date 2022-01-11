package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"spannerapp/internal/credentials"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

func App(creds credentials.SpannerCredentials) *mux.Router {
	client, err := connect(creds)
	if err != nil {
		log.Printf(" Error creating client: %s", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{key}", handleSet(*client)).Methods("PUT")
	r.HandleFunc("/{key}", handleGet(*client)).Methods("GET")

	return r
}

func connect(creds credentials.SpannerCredentials) (*spanner.Client, error) {
	ctx := context.Background()

	client, err := spanner.NewClient(context.Background(), creds.FullDBName, option.WithCredentialsJSON([]byte(creds.Credentials)))
	stmt := spanner.Statement{SQL: `SELECT count(*) As tableCount
	                                FROM INFORMATION_SCHEMA.TABLES
                                    WHERE TABLE_NAME = 'test'`}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return client, nil
	}
	if err != nil {
		return nil, err
	}
	var tableCount int64
	if err := row.Columns(&tableCount); err != nil {
		return nil, err
	}
	if tableCount != 0 {
		return client, nil
	}

	adminClient, err := database.NewDatabaseAdminClient(ctx, option.WithCredentialsJSON([]byte(creds.Credentials)))
	if err != nil {
		log.Printf(" Error creating admin client: %s", err)
		return nil, err
	}
	defer adminClient.Close()

	op, err := adminClient.UpdateDatabaseDdl(ctx, &adminpb.UpdateDatabaseDdlRequest{
		Database: creds.FullDBName,
		Statements: []string{
			`CREATE TABLE test (
				keyname  STRING(1024),
				valuedata   STRING(1024),
			) PRIMARY KEY (keyname)`,
		},
	})
	if err != nil {
		log.Printf(" Error creating table 1: %s", err)
		return nil, err
	}
	if err := op.Wait(ctx); err != nil {
		log.Printf(" Error creating table 2: %s", err)
		return nil, err
	}

	return client, err
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
