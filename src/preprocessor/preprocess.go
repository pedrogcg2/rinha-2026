package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pedrogcg2/rinha-2026/index"
)

const (
	maxSize = 3_000_000
)

type transactionReference struct {
	Vector []float64 `json:"vector"`
	Legit  string    `json:"label"`
}

func main() {
	Process("../../resources/references.json.gz")
}

func Process(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	g, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(g)
	if _, err = dec.Token(); err != nil {
		panic(err)
	}
	current := transactionReference{Vector: make([]float64, 0, 14), Legit: ""}
	t := make([]*index.QuantizeTransaction, maxSize)
	c := 0

	for dec.More() {
		if c == maxSize {
			break
		}
		err = dec.Decode(&current)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err.Error())
			panic(err)
		}
		t[c] = getTransactions(current)
		c++
	}

	tree := index.BuildVpTree(t)

	saveTree(tree)
}

func saveTree(tree *index.VpTree) {
	file, err := os.Create("../../resources/idx.bin")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, tree.Nodes)
	if err != nil {
		panic(err)
	}
}

func getTransactions(r transactionReference) *index.QuantizeTransaction {
	v := [14]int16{}
	for i, t := range r.Vector {
		v[i] = index.Quantize(t)
	}
	return &index.QuantizeTransaction{Vector: v, Legit: r.Legit == "legit"}
}
