package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"spannerapp/internal/credentials"
	"strings"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "data"
)

func App(creds credentials.SpannerCredentials) http.HandlerFunc {
	client, err := connect(creds)
	if err != nil {
		log.Printf(" Error creating client: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleGet(w, key, client)
		case http.MethodPut:
			handleSet(w, r, key, client)
		default:
			fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
	}
}

func connect(creds credentials.SpannerCredentials) (*spanner.Client, error) {
	ctx := context.Background()

	client, err := spanner.NewClient(ctx, creds.FullDBName, option.WithCredentialsJSON([]byte(creds.Credentials)))
	stmt := spanner.Statement{
		SQL: `SELECT count(*) As tableCount
	                                FROM INFORMATION_SCHEMA.TABLES
                                    WHERE TABLE_NAME = @tableName`,
		Params: map[string]any{
			"tableName": tableName,
		}}

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
			fmt.Sprintf(`CREATE TABLE %s (
				%s  STRING(1024),
				%s   STRING(1024),
			) PRIMARY KEY (%s)`, tableName, keyColumn, valueColumn, keyColumn),
		},
	})
	if err != nil {
		log.Printf("Error creating table: %s", err)
		return nil, err
	}
	if err := op.Wait(ctx); err != nil {
		log.Printf("Error waiting for table creation: %s", err)
		return nil, err
	}

	return client, err
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
