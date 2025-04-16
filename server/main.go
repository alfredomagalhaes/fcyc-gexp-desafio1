package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Currence struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Bid  string `json:"bid"`
}

const databaseFile string = "quotations.db"
const createScript string = `
  CREATE TABLE IF NOT EXISTS dolar_prices (
  id INTEGER NOT NULL PRIMARY KEY,
  price double NOT NULL,
  time DATETIME NOT NULL
  );`

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", databaseFile)

	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	err = setupDatabase()

	if err != nil {
		log.Fatalf("failed to create table at sqlite database: %v", err)
	}

	http.HandleFunc("/cotacao", getDollarPriceHandler)
	http.ListenAndServe(":8080", nil)
}

func getDollarPriceHandler(w http.ResponseWriter, r *http.Request) {
	ctxReq, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		if ctxReq.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			fmt.Println("Request to get Dollar price timed out")
			return
		}
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch data", response.StatusCode)
		return
	}
	dollarResponse, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	price := getPrice(dollarResponse)
	savePriceToDb(price)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dollarResponse)

}
func getPrice(buffer []byte) float64 {
	var quoteMap map[string]Currence
	var price float64
	err := json.Unmarshal(buffer, &quoteMap)
	if err != nil {
		return 0.00
	}
	price, _ = strconv.ParseFloat(quoteMap["USDBRL"].Bid, 64)
	return price
}

func setupDatabase() error {
	_, err := db.Exec(createScript)
	return err

}
func savePriceToDb(price float64) error {
	ctxDb, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := db.ExecContext(ctxDb, "INSERT INTO dolar_prices VALUES(NULL,?,?)", price, time.Now())

	return err
}
