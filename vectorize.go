package main

import (
	"slices"
	"time"
)

type NormalizationConstants struct {
	MaxAmount                       float32 `json:"max_amount"`
	MaxInstallments                 float32 `json:"max_installments"`
	AmountVsAverageRatio            float32 `json:"amount_vs_avg_ratio"`
	MaxMinutes                      float32 `json:"max_minutes"`
	MaxKm                           float32 `json:"max_km"`
	MaxTransactionsCountLast24hours float32 `json:"max_tx_count_24h"`
	MaxMerchantAverageAmount        float32 `json:"max_merchant_avg_amount"`
}

var MCC_RISK map[string]float32
var Constants NormalizationConstants

func VectorizeTransaction(transaction *TransactionRequest) [14]float32 {
	var vector [14]float32

	vector[0] = clamp(transaction.Transaction.Amount / Constants.MaxAmount)
	vector[1] = clamp(transaction.Transaction.Installments / Constants.MaxInstallments)
	vector[2] = clamp((transaction.Transaction.Amount / transaction.Customer.AverageAmount) / Constants.AmountVsAverageRatio)
	vector[3] = clamp(float32(transaction.Transaction.RequestedAt.Hour()) / 23)
	vector[4] = clamp(GetWeekDay(transaction.Transaction.RequestedAt) / 6)
	vector[5] = -1
	vector[6] = -1
	if transaction.LastTransaction != nil {
		vector[5] = clamp(float32(transaction.Transaction.RequestedAt.Sub(transaction.LastTransaction.TimeStamp).Minutes() / float64(Constants.MaxMinutes)))
		vector[6] = clamp(transaction.LastTransaction.KmFromCurrent / Constants.MaxKm)
	}
	vector[7] = clamp(transaction.Terminal.KmFromHome / Constants.MaxKm)
	vector[8] = clamp(transaction.Customer.LastDayHoursTransactionsCount / Constants.MaxTransactionsCountLast24hours)
	vector[9] = clampBool(transaction.Terminal.IsOnline)
	vector[10] = clampBool(transaction.Terminal.CardPresent)
	vector[11] = clampBool(!slices.Contains(transaction.Customer.KnownMerchants, transaction.Merchant.Id))
	vector[12] = 0.5
	risk, exists := MCC_RISK[transaction.Merchant.Mcc]
	if exists {
		vector[12] = risk
	}
	vector[13] = clamp(transaction.Merchant.AverageAmount / Constants.MaxMerchantAverageAmount)
	return vector
}

func clampBool(b bool) float32 {
	if b {
		return 1.0
	}
	return 0.0
}

func GetWeekDay(t time.Time) float32 {
	day := float32(t.Weekday())
	if day == 0 {
		return 6.0
	}
	return day - 1
}

func clamp(value float32) float32 {
	if value > 1 {
		return 1.0
	}

	if value < 0 {
		return 0.0
	}

	return value
}
