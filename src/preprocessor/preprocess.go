package preprocessor

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"time"

	"gonum.org/v1/gonum/mat"
)

type transaction struct {
	embedding *mat.VecDense
	legit     bool
}

type transactionReference struct {
	Vector []float64 `json:"vector"`
	Legit  bool      `json:"legit"`
}

func Process(path string) *VpTreeNode {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	transactionsRaw := []transactionReference{}
	json.NewDecoder(file).Decode(&transactionsRaw)
	transactions := getTransactionsLabel(transactionsRaw[:])

	//TODO: to desalocando memoria inutilizada para evitar estouro
	//quando de fato tiver um preprocessamento decente, talvez isso nao seja necessario.
	//nao custa nada manter (eu acho)
	file.Close()
	transactionsRaw = nil
	log.Println("Calling GC:")
	time.Sleep(time.Second * 3)
	runtime.GC()
	return buildVpTree(transactions)
}

func ProcessHeap(path string) *Heap {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	transactionsRaw := []transactionReference{}
	json.NewDecoder(file).Decode(&transactionsRaw)
	transactions := getTransactionsLabel(transactionsRaw[:])

	//TODO: to desalocando memoria inutilizada para evitar estouro
	//quando de fato tiver um preprocessamento decente, talvez isso nao seja necessario.
	//nao custa nada manter (eu acho)
	file.Close()
	transactionsRaw = nil
	log.Println("Calling GC:")
	time.Sleep(time.Second * 3)
	runtime.GC()
	return &Heap{transactions: transactions}
}

func getTransactionsLabel(r []transactionReference) []transaction {
	result := make([]transaction, 0, len(r))

	for _, t := range r {
		result = append(result, transaction{legit: t.Legit, embedding: mat.NewVecDense(len(t.Vector), t.Vector)})
	}

	return result
}
