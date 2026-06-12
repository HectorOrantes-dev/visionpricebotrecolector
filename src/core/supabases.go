package core

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitPostgres() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		// Por defecto lo ponemos en "require" porque las bases de datos en la nube (como Supabase, AWS, Render) exigen SSL.
		// Si usas una base de datos local, puedes agregar DB_SSLMODE=disable en tu .env
		sslmode = "require"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	log.Println("PostgreSQL connected successfully")
}
