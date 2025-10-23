package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./monitoring.db")
	if err != nil {
		return err
	}

	// Crear tabla para métricas
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		total_ram INTEGER,
		free_ram INTEGER,
		container_count INTEGER,
		containers_eliminated INTEGER DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS container_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		container_id TEXT,
		pid INTEGER,
		name TEXT,
		memory_usage REAL,
		cpu_usage REAL,
		status TEXT
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	fmt.Println("Base de datos inicializada")
	return nil
}

func saveMetrics(totalRAM, freeRAM, containerCount int) error {
	_, err := db.Exec(
		"INSERT INTO metrics (total_ram, free_ram, container_count) VALUES (?, ?, ?)",
		totalRAM, freeRAM, containerCount,
	)
	return err
}
