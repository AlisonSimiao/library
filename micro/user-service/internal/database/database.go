package database

import (
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
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

func RunMigrations(db *sqlx.DB, migrationsDir string) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Erro ao ler diretório de migrations: %v", err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}
		path := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Erro ao ler migration %s: %v", file.Name(), err)
		}
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Erro ao executar migration %s: %v", file.Name(), err)
		}
		log.Printf("Migration executada: %s", file.Name())
	}
}
