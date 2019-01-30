package db

import (
	"fmt"
	"testing"
)

func (db *Database) TestReset(t *testing.T) {
	_, err := db.sqlDB.Exec("DELETE FROM orders")
	if err != nil {
		t.Fatalf("DELETE FROM orders err = %v; want nil", err)
	}
	_, err = db.sqlDB.Exec("DELETE FROM campaigns")
	if err != nil {
		t.Fatalf("DELETE FROM campaigns err = %v; want nil", err)
	}
}

func (db *Database) TestCount(t *testing.T, table string) int {
	var n int
	err := db.sqlDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
	if err != nil {
		t.Fatalf("Scan() err = %v; want nil", err)
	}
	return n
}
