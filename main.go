package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("Configura DATABASE_URL para ejecutar el servidor de la raíz")
	}

	dbConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("DATABASE_URL no es válida para PostgreSQL")
	}
	db := stdlib.OpenDB(*dbConfig)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"estado":"vivo"}`)
	})

	http.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
		var version string

		if err := db.QueryRow("select version()").Scan(&version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `{"bd":"sin conexion","detalle":"No se pudo consultar PostgreSQL"}`)
			return
		}

		fmt.Fprintf(w, `{"bd":"ok","version":%q}`, version)
	})

	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
