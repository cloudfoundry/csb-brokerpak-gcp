package app

import (
	"fmt"
	"log"
	"mysqlapp/internal/connector"
	"net/http"
	"strings"
)

const (
	tlsQueryParam = "tls"
)

func App(conn connector.Connector) http.Handler {

	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && strings.Trim(r.URL.Path, "/") == "":
			aliveness(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	r.HandleFunc("/key-value/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, key, conn)
		case http.MethodPut:
			handleSet(w, r, key, conn)
		default:
			methodNotAllowed(w)
		}
	})

	r.HandleFunc("/admin/ssl/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			handleGetSSLCipher(w, r, conn)
		default:
			methodNotAllowed(w)
		}
	})

	return r
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

func methodNotAllowed(w http.ResponseWriter) {
	fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}
