package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/ReviewExplorer"
	db, err := sql.Open("mysql", dsn)
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		log.Fatalf("Ошибка подключения бд", err)
	}
	defer db.Close()

	f, _ := os.Open("data/data-2263-2025-12-29.csv")
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	rows, _ := reader.ReadAll()

	stmt, _ := db.Prepare("INSERT INTO schools (org_type, full_name, short_name, legal_form, address, website) VALUES (?,?,?,?,?,?)")
	defer stmt.Close()

	count := 0
	for i, row := range rows {
		if i == 0 || len(row) < 15 {
			continue
		}
		_, err := stmt.Exec(row[0], row[3], row[4], row[4], row[12], row[16])
		if err == nil {
			count++
		}
	}

	fmt.Printf("итого - %v", count)
}
