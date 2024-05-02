package main

import (
	"encoding/csv"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Record struct {
	Name  string
	Score int
}

func main() {
	// インスタンスを作成
	e := echo.New()

	// ミドルウェアを設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080", "https://takaryo1010.github.io"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	// ルートを設定
	e.POST("/write", writeCSV)
	e.GET("/read", readCSV)

	e.Logger.Fatal(e.Start(":8080"))
}

// ハンドラーを定義
func writeCSV(c echo.Context) error {
	name := c.FormValue("name")
	score := c.FormValue("score")
	// Open CSV file in append mode
	file, err := os.OpenFile("ranking.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	records := [][]string{
		[]string{name, score},
		// Add more records here if needed
	}

	w := csv.NewWriter(file)
	if err := w.WriteAll(records); err != nil {
		return err
	}

	// Return success response
	return c.String(http.StatusOK, "CSV file updated successfully")
}

func readCSV(c echo.Context) error {
	// Open CSV file
	file, err := os.Open("ranking.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	// Read CSV records
	r := csv.NewReader(file)
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}

	// Convert CSV records to struct and map for deduplication
	nameMap := make(map[string]int) // Map to track name-score pairs
	var records []Record
	for _, row := range rows {
		score, err := strconv.Atoi(row[1])
		if err != nil {
			return err
		}
		// Check if the name already exists in the map
		if existingScore, ok := nameMap[row[0]]; ok {
			// If the current score is greater than the existing score, update the map
			if score > existingScore {
				nameMap[row[0]] = score
			}
		} else {
			nameMap[row[0]] = score
		}
	}

	// Convert map to slice of Record
	for name, score := range nameMap {
		records = append(records, Record{Name: name, Score: score})
	}

	// Sorting records by score in descending order
	sort.Slice(records, func(i, j int) bool {
		return records[i].Score > records[j].Score
	})

	// Limiting to top 10 records
	topRecords := records
	if len(records) > 10 {
		topRecords = records[:10]
	}

	// Write topRecords to CSV file
	file, err = os.Create("ranking.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)

	// Write sorted records to CSV
	var stringRecords [][]string
	for _, record := range topRecords {
		stringRecords = append(stringRecords, []string{record.Name, strconv.Itoa(record.Score)})
	}
	if err := w.WriteAll(stringRecords); err != nil {
		return err
	}
	w.Flush()

	// Return JSON data
	return c.JSON(http.StatusOK, topRecords)
}
