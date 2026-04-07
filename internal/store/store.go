package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Reservations struct {
	ID string `json:"id"`
	GuestName string `json:"guest_name"`
	GuestPhone string `json:"guest_phone"`
	GuestEmail string `json:"guest_email"`
	PartySize int64 `json:"party_size"`
	Date string `json:"date"`
	Time string `json:"time"`
	Table string `json:"table"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "reservation.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS reservations(id TEXT PRIMARY KEY, guest_name TEXT NOT NULL, guest_phone TEXT DEFAULT '', guest_email TEXT DEFAULT '', party_size INTEGER NOT NULL, date TEXT NOT NULL, time TEXT NOT NULL, table TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL, record_id TEXT NOT NULL, data TEXT NOT NULL DEFAULT '{}', PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateReservations(e *Reservations) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO reservations(id, guest_name, guest_phone, guest_email, party_size, date, time, table, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.GuestName, e.GuestPhone, e.GuestEmail, e.PartySize, e.Date, e.Time, e.Table, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetReservations(id string) *Reservations {
	var e Reservations
	if d.db.QueryRow(`SELECT id, guest_name, guest_phone, guest_email, party_size, date, time, table, status, notes, created_at FROM reservations WHERE id=?`, id).Scan(&e.ID, &e.GuestName, &e.GuestPhone, &e.GuestEmail, &e.PartySize, &e.Date, &e.Time, &e.Table, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListReservations() []Reservations {
	rows, _ := d.db.Query(`SELECT id, guest_name, guest_phone, guest_email, party_size, date, time, table, status, notes, created_at FROM reservations ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Reservations
	for rows.Next() { var e Reservations; rows.Scan(&e.ID, &e.GuestName, &e.GuestPhone, &e.GuestEmail, &e.PartySize, &e.Date, &e.Time, &e.Table, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateReservations(e *Reservations) error {
	_, err := d.db.Exec(`UPDATE reservations SET guest_name=?, guest_phone=?, guest_email=?, party_size=?, date=?, time=?, table=?, status=?, notes=? WHERE id=?`, e.GuestName, e.GuestPhone, e.GuestEmail, e.PartySize, e.Date, e.Time, e.Table, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteReservations(id string) error {
	_, err := d.db.Exec(`DELETE FROM reservations WHERE id=?`, id)
	return err
}

func (d *DB) CountReservations() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM reservations`).Scan(&n); return n
}

func (d *DB) SearchReservations(q string, filters map[string]string) []Reservations {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (guest_name LIKE ? OR guest_phone LIKE ? OR guest_email LIKE ? OR time LIKE ? OR table LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, guest_name, guest_phone, guest_email, party_size, date, time, table, status, notes, created_at FROM reservations WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Reservations
	for rows.Next() { var e Reservations; rows.Scan(&e.ID, &e.GuestName, &e.GuestPhone, &e.GuestEmail, &e.PartySize, &e.Date, &e.Time, &e.Table, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

// GetExtras returns the JSON extras blob for a record. Returns "{}" if none.
func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(`SELECT data FROM extras WHERE resource=? AND record_id=?`, resource, recordID).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

// SetExtras stores the JSON extras blob for a record.
func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?) ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`, resource, recordID, data)
	return err
}

// DeleteExtras removes extras when a record is deleted.
func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(`DELETE FROM extras WHERE resource=? AND record_id=?`, resource, recordID)
	return err
}

// AllExtras returns all extras for a resource type as a map of record_id → JSON string.
func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(`SELECT record_id, data FROM extras WHERE resource=?`, resource)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
