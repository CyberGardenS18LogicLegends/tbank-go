package sqlite

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func InitializeDatabase(dbPath string, log *slog.Logger) (*sql.DB, error) {
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)

	// Open or create the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error("failed to open database", slog.String("dbPath", dbPath), slog.Any("error", err))
		return nil, err
	}

	if !dbExists {
		err = createTables(db, log)
		if err != nil {
			log.Error("failed to create tables", slog.String("dbPath", dbPath), slog.Any("error", err))
			return nil, err
		}
		log.Info("Database and tables initialized successfully", slog.String("dbPath", dbPath))
	} else {
		log.Info("Database file already exists", slog.String("dbPath", dbPath))
	}

	return db, nil
}

func createTables(db *sql.DB, log *slog.Logger) error {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    registered_at TEXT NOT NULL
);`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Error("failed to create users table", slog.Any("error", err))
		return err
	}

	log.Info("Users table created successfully")
	return nil
}
