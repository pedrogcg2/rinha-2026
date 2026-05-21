package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pedrogcg2/rinha-2026/index"
)

var (
	vpTree     *index.VpTree = nil
	dataLoaded bool          = false
	responses  [6][]byte
)

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
	initResponses()
	LoadVpTreeFromBin("../resources/idx.bin")
	dataLoaded = true
}

func initResponses() {
	for i := range 6 {
		approved := i < 3
		value := float64(i / 10)
		b, err := json.Marshal(FraudScoreResponse{Approved: approved, FraudScore: value})
		if err != nil {
			panic(err)
		}
		responses[i] = b
	}
}

func LoadVpTreeFromBin(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	vpTree = &index.VpTree{}
	bf := bufio.NewReader(f)
	for j := range 3_000_000 {

		if j%500_000 == 0 {
			runtime.GC()
		}
		err = binary.Read(bf, binary.LittleEndian, &vpTree.Nodes[j])
		if err != nil {
			panic(err)
		}
	}
	runtime.GC()
}

func addEndpoints() {
	http.HandleFunc("GET /ready", ready)
	http.HandleFunc("POST /fraud-score", fraudScore)
}

func ready(w http.ResponseWriter, r *http.Request) {
	if !dataLoaded {
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
}

func fraudScore(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	payload := TransactionRequest{}
	decoder.Decode(&payload)
	score := calculateFraudScore(&payload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responses[score])
}

func calculateFraudScore(r *TransactionRequest) int8 {
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
