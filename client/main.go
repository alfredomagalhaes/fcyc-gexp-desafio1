package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"log"
)

type CurrencyResponse struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Error while preparing request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Request timed out")
			return
		}
		log.Fatalf("Error while preparing request: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get dollar price: %v", err)
	}
	var currencyResponse CurrencyResponse
	//io.Copy(os.Stdout, res.Body)
	err = json.NewDecoder(res.Body).Decode(&currencyResponse)
	if err != nil {
		log.Fatalf("Error while decoding response: %v", err)
	}

	textToSave := fmt.Sprintf("Dolar: %f", currencyResponse.Bid)
	err = os.WriteFile("cotacao.txt", []byte(textToSave), 0644)
	if err != nil {
		log.Fatalf("Error while opening file: %v", err)
	}

}
