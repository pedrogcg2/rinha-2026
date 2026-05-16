package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pedrogcg2/rinha-2026/preprocessor"
)

var vpTree *preprocessor.VpTreeNode

func main() {
	addEndpoints()
	initConstants()
	server := &http.Server{Addr: ":8080", ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second}
	server.ListenAndServe()
}

func initConstants() {
	mccRiskBuff, err := os.ReadFile("../resources/mcc_risk.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(mccRiskBuff, &MCC_RISK)
	if err != nil {
		panic(err)
	}
	normalizationConstants, err := os.ReadFile("../resources/normalization.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(normalizationConstants, &Constants)
	if err != nil {
		panic(err)
	}

	vpTree = preprocessor.Process("../resources/references.json.gz")
}

func addEndpoints() {
	http.HandleFunc("GET /ready", ready)
	http.HandleFunc("POST /fraud-score", fraudScore)
}

func ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func fraudScore(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var payload TransactionRequest

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&payload)
	score := calculate_fraud_score(&payload)
	approved := score < 0.6
	response := FraudScoreResponse{Approved: approved, FraudScore: score}
	json.NewEncoder(w).Encode(response)
	elapsed := time.Since(start)
	log.Printf("Elapsed in %d ms", elapsed.Milliseconds())
}

func calculate_fraud_score(r *TransactionRequest) float64 {
	v := VectorizeTransaction(r)
	return vpTree.Query(v)
}

type FraudScoreResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float64 `json:"fraud_score"`
}

type TransactionRequest struct {
	Transaction     TransactionInfo      `json:"transaction"`
	Customer        CustomerInfo         `json:"customer"`
	Merchant        MerchantInfo         `json:"merchant"`
	Terminal        TerminalInfo         `json:"terminal"`
	LastTransaction *LastTransactionInfo `json:"last_transaction"`
}

type TransactionInfo struct {
	Amount       float64   `json:"amount"`
	Installments float64   `json:"installments"`
	RequestedAt  time.Time `json:"requested_at"`
}

type CustomerInfo struct {
	AverageAmount                 float64  `json:"avg_amount"`
	LastDayHoursTransactionsCount float64  `json:"tx_count_24h"`
	KnownMerchants                []string `json:"known_merchants"`
}

type MerchantInfo struct {
	Id            string  `json:"id"`
	Mcc           string  `json:"mcc"`
	AverageAmount float64 `json:"avg_amount"`
}

type TerminalInfo struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool    `json:"card_present"`
	KmFromHome  float64 `json:"km_from_home"`
}

type LastTransactionInfo struct {
	TimeStamp     time.Time `json:"timestamp"`
	KmFromCurrent float64   `json:"km_from_current"`
}
