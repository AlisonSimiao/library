package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) *sqlx.DB {
	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de dados: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Erro ao fazer ping no banco de dados: %v", err)
	}

	log.Println("Conectado ao banco de dados com sucesso")
	return db
}
