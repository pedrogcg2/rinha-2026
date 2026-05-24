package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/pedrogcg2/rinha-2026/index"
	"github.com/valyala/fasthttp"
)

var (
	vpTree     *index.VpTree = nil
	dataLoaded bool          = false
	responses  [6][]byte
)

func main() {
	initConstants()
	initResponses()

	fasthttp.ListenAndServe(":8080", handler)
}

func initConstants() {
	mccRiskBuff, err := os.ReadFile("../resources/mcc_risk.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(mccRiskBuff, &MccRisk)
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

func handler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	switch {
	case path == "/fraud-score":
		fraudScore(ctx)

	case path == "/ready":
		ready(ctx)

	}
}

func ready(ctx *fasthttp.RequestCtx) {
	if !dataLoaded {
		ctx.Response.Header.SetStatusCode(401)
		return
	}
	ctx.Response.Header.SetStatusCode(200)
}

func fraudScore(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	payload := TransactionRequest{}
	json.Unmarshal(body, &payload)
	score := calculateFraudScore(&payload)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.AppendBody(responses[score])
}

func calculateFraudScore(r *TransactionRequest) int8 {
	v := VectorizeTransaction(r)
	result := vpTree.Query(v)
	return result
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
