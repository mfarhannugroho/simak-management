package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"simak-backend/config"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down|version|force <version>]")
		os.Exit(1)
	}

	command := args[0]
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("✅ Database migration applied successfully!")
	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("✅ Database migration rollback (1 step) completed!")
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to fetch migration version: %v", err)
		}
		log.Printf("Current migration version: %d (dirty: %v)\n", v, dirty)
	case "force":
		if len(args) < 2 {
			log.Fatal("Version required for force command, e.g. force 3")
		}
		var targetVersion int
		if _, err := fmt.Sscanf(args[1], "%d", &targetVersion); err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(targetVersion); err != nil {
			log.Fatalf("Force migration version failed: %v", err)
		}
		log.Printf("✅ Migration forced to version %d\n", targetVersion)
	default:
		log.Fatalf("Unknown command: %s. Use up, down, version, or force <version>.", command)
	}
}
