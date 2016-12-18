package main

import (
	"encoding/json"
	"fmt"
	"github.com/linxlunx/gostemmer"
)

type Response struct {
	Word  string                       `json:"word"`
	Count int                          `json:"count"`
	Roots map[string]map[string]string `json:"roots"`
}

func main() {
	// Set lokasi kamus terlebih dahulu
	dict := "../kamus.txt"

	word := "melihat"
	rooted := gostemmer.StemWord(word, dict)

	for key, value := range rooted {
		temp := &Response{
			Word:  key,
			Count: value.Count,
			Roots: value.Roots,
		}

		resp, _ := json.Marshal(temp)
		fmt.Println(string(resp))
	}

}
