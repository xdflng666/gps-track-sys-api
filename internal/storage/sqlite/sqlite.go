package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gps-track-sys-api/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const funcName = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", funcName, err)
	}

	// Ugly table&index creation on storage initialization
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS Roles (
		id INTEGER PRIMARY KEY,
		role TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS Users (
		id INTEGER PRIMARY KEY,
		login TEXT NOT NULL,
		password TEXT NOT NULL,
		role_id INTEGER,
		FOREIGN KEY (role_id) REFERENCES Roles(id)
	);
	
	CREATE TABLE IF NOT EXISTS Devices (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS PositionData (
		id INTEGER PRIMARY KEY,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		timestamp TEXT NOT NULL,
		device_id INTEGER,
		FOREIGN KEY (device_id) REFERENCES Devices(id)
	);
	
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", funcName, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", funcName, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) AddDevice(name string, description string) error {
	const funcName = "storage.sqlite.AddDevice"

	stmt, err := s.db.Prepare("INSERT INTO Devices(name, description) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", funcName, err)
	}

	_, err = stmt.Exec(name, description)
	if err != nil {
		return fmt.Errorf("%s: %w", funcName, err)
	}

	return nil
}

func (s *Storage) DeleteDevice(deviceID int64) (int64, error) {
	const funcName = "storage.sqlite.DeleteDevice"

	stmt, err := s.db.Prepare("DELETE FROM Devices WHERE id = ?")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", funcName, err)
	}

	res, err := stmt.Exec(deviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", funcName, storage.ErrDeviceNotFound)
		}
		return 0, fmt.Errorf("%s: execute statement: %w", funcName, err)
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", funcName, err)
	}

	return rowsCount, nil
}
