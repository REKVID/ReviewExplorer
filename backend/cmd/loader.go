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
		log.Fatalf("Ошибка подключения бд %v", err)
	}
	defer db.Close()

	f, _ := os.Open("data/data-2263-2025-12-29.csv")
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	rows, _ := reader.ReadAll()

	stmt, _ := db.Prepare("INSERT IGNORE INTO schools (org_type, full_name, short_name, legal_form, address, website, lat, lon) VALUES (?,?,?,?,?,?,?,?)")
	defer stmt.Close()

	count := 0
	for i, row := range rows {
		if i == 0 || len(row) < 26 {
			continue
		}

		// Парсим координаты из поля geodata_center (индекс 25)
		// Формат: "{coordinates=[37.644301812, 55.813793091], type=Point}"
		geo := row[25]
		var lat, lon float64
		// Простое извлечение чисел между скобками [ ]
		// CSV Go не парсит вложенный JSON, поэтому делаем примитивный парсинг строки
		inCoords := false
		coordStr := ""

		for j := 0; j < len(geo); j++ {
			if geo[j] == '[' {
				inCoords = true
				continue
			}
			if geo[j] == ']' {
				inCoords = false
				break
			}
			if inCoords {
				coordStr += string(geo[j])
			}
		}

		if coordStr != "" {
			fmt.Sscanf(coordStr, "%f, %f", &lon, &lat) // В GeoJSON сначала долгота, потом широта!
		}

		_, err := stmt.Exec(row[0], row[3], row[4], row[9], row[12], row[16], lat, lon)
		if err == nil {
			count++
		} else {
			// fmt.Printf("Error inserting row %d: %v\n", i, err)
		}
	}

	fmt.Printf("Успешно загружено школ: %v\n", count)
}
