// Sirve esta carpeta en http://localhost:5173, el origen que su API permite.
// Se corre desde esta carpeta con: go run servir.go
package main

import (
	"log"
	"net"
	"net/http"
)

func main() {
	escucha, err := net.Listen("tcp", "localhost:5173")
	if err != nil {
		log.Fatal(err) // casi siempre: otro programa ya usa el 5173
	}
	log.Println("abra http://localhost:5173 en el navegador (Ctrl+C para terminar)")
	log.Fatal(http.Serve(escucha, http.FileServer(http.Dir("."))))
}
