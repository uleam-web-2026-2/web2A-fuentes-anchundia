package main

import (
	"errors"
	"fidelizacion/internal/fidelizacion"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	if configuracion := os.Getenv("ENV_FILE"); configuracion != "" {
		if err := godotenv.Load(configuracion); err != nil {
			log.Fatal("No se pudo leer ENV_FILE")
		}
	} else {
		_ = godotenv.Load()
	}
	duracion := time.Hour
	if valor := os.Getenv("JWT_TTL"); valor != "" {
		var err error
		duracion, err = time.ParseDuration(valor)
		if err != nil || duracion <= 0 {
			log.Fatal("JWT_TTL debe ser una duración positiva, por ejemplo 1h")
		}
	}
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	numero, err := strconv.Atoi(puerto)
	if err != nil || numero < 1 || numero > 65535 {
		log.Fatal("PORT debe estar entre 1 y 65535")
	}
	db, err := fidelizacion.Conectar()
	if err != nil {
		log.Fatalf("No se pudo conectar a la base: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	auth, err := fidelizacion.NuevaAutenticacion(db, os.Getenv("JWT_SECRET"), duracion)
	if err != nil {
		log.Fatal(err)
	}
	if err := fidelizacion.PrepararBase(db); err != nil {
		log.Fatalf("No se pudo actualizar la base conservando los datos: %v", err)
	}
	log.Printf("PostgreSQL: base %s en %s:%s", os.Getenv("DB_NAME"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	modo := os.Getenv("GIN_MODE")
	if modo == "" {
		modo = gin.ReleaseMode
	}
	gin.SetMode(modo)
	r := gin.Default()
	_ = r.SetTrustedProxies(nil)
	r.Use(func(c *gin.Context) { c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20); c.Next() })
	// El servicio y el transporte solo reciben el puerto del negocio.
	var repo fidelizacion.Repositorio
	switch adaptador := os.Getenv("REPOSITORY_ADAPTER"); adaptador {
	case "", "sqlc":
		repo = fidelizacion.NuevoRepositorioSqlc(&fidelizacion.RepositorioGORM{DB: db}, sqlDB)
	case "gorm":
		repo = &fidelizacion.RepositorioGORM{DB: db}
	default:
		log.Fatal("REPOSITORY_ADAPTER debe ser sqlc o gorm")
	}
	fidelizacion.ConfigurarRutasConRepositorio(r, repo, auth)
	servidor := &http.Server{Addr: "127.0.0.1:" + puerto, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("LealCaf listo en http://localhost:%s", puerto)
	if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
