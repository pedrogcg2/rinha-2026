package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	addEndpoints()
	initConstants()
	server := &http.Server{Addr: ":9999", ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	server.ListenAndServe()
}

func initConstants() {
	mccRiskBuff, err := os.ReadFile("mcc_risk.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(mccRiskBuff, &MCC_RISK)
	if err != nil {
		panic(err)
	}
	normalizationConstants, err := os.ReadFile("normalization.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(normalizationConstants, &Constants)
	if err != nil {
		panic(err)
	}
}

func addEndpoints() {
	http.HandleFunc("GET /ready", ready)
	http.HandleFunc("POST /fraud-score", fraudScore)
}

func ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func fraudScore(w http.ResponseWriter, r *http.Request) {
	var payload TransactionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&payload)
	score := calculate_fraud_score(&payload)
	approved := score < 0.6
	response := FraudScoreResponse{Approved: approved, FraudScore: score}
	json.NewEncoder(w).Encode(response)
}

func calculate_fraud_score(r *TransactionRequest) float32 {
	v := VectorizeTransaction(r)
	fmt.Println(v)
	return 0.5
}

type FraudScoreResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float32 `json:"fraud_score"`
}

type TransactionRequest struct {
	Transaction     TransactionInfo      `json:"transaction"`
	Customer        CustomerInfo         `json:"customer"`
	Merchant        MerchantInfo         `json:"merchant"`
	Terminal        TerminalInfo         `json:"terminal"`
	LastTransaction *LastTransactionInfo `json:"last_transaction"`
}

type TransactionInfo struct {
	Amount       float32   `json:"amount"`
	Installments float32   `json:"installments"`
	RequestedAt  time.Time `json:"requested_at"`
}

type CustomerInfo struct {
	AverageAmount                 float32  `json:"avg_amount"`
	LastDayHoursTransactionsCount float32  `json:"tx_count_24h"`
	KnownMerchants                []string `json:"known_merchants"`
}

type MerchantInfo struct {
	Id            string  `json:"id"`
	Mcc           string  `json:"mcc"`
	AverageAmount float32 `json:"avg_amount"`
}

type TerminalInfo struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool    `json:"card_present"`
	KmFromHome  float32 `json:"km_from_home"`
}

type LastTransactionInfo struct {
	TimeStamp     time.Time `json:"timestamp"`
	KmFromCurrent float32   `json:"km_from_current"`
}
