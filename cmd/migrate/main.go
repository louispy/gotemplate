package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/louispy/template/internal/database"
)

type Config struct {
	DB     database.Config `envPrefix:"DB_"`
	SqlDir string          `env:"SQL_DIR" envDefault:"./sql"`
}

func main() {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("cannot parse config: %v", err)
	}

	ctx := context.Background()
	db, err := database.New(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer db.Close()

	files, err := filepath.Glob(filepath.Join(cfg.SqlDir, "*.sql"))
	if err != nil {
		log.Fatalf("cannot list sql files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no sql files found in %s", cfg.SqlDir)
	}
	sort.Strings(files)

	for _, f := range files {
		name := filepath.Base(f)
		log.Printf("running %s", name)
		content, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			log.Fatalf("exec %s: %v", name, err)
		}
	}
	log.Println("migration complete")
}
