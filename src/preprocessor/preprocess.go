package preprocessor

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math"
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
	Legit  string    `json:"label"`
}

func Process(path string) *VpTreeNode {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}

	transactionsRaw := []transactionReference{}
	json.NewDecoder(r).Decode(&transactionsRaw)
	fmt.Println(len(transactionsRaw))
	transactions := getTransactionsLabel(transactionsRaw[:int(2.3*math.Pow10(5))])

	//TODO: to desalocando memoria inutilizada para evitar estouro
	//quando de fato tiver um preprocessamento decente, talvez isso nao seja necessario.
	//nao custa nada manter (eu acho)
	r.Close()

	transactionsRaw = nil
	log.Println("Calling GC:")
	time.Sleep(time.Second * 3)
	runtime.GC()
	return buildVpTree(transactions)
}

func getTransactionsLabel(r []transactionReference) []transaction {
	result := make([]transaction, 0, len(r))

	for _, t := range r {
		legit := t.Legit == "legit"
		result = append(result, transaction{legit: legit, embedding: mat.NewVecDense(len(t.Vector), t.Vector)})
	}

	return result
}
