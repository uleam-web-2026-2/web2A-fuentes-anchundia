package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx",
		"postgres://postgres:SU_CONTRASEÑA@localhost:5432/mesa_ayuda_web2")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"estado":"vivo"}`)
	})

	http.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
		var version string

		if err := db.QueryRow("select version()").Scan(&version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"bd":"sin conexion","detalle":%q}`, err.Error())
			return
		}

		fmt.Fprintf(w, `{"bd":"ok","version":%q}`, version)
	})

	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}