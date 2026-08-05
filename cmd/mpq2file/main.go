package main

import (
	"fmt"
	"log"

	"github.com/gravestench/mpq"
)

func main() {
	mpqHandle, err := mpq.New("/Users/dylanknuth/d2_english_mpq/patch_d2.mpq")
	if err != nil {
		log.Fatalf("opening mpq archive: %v", err)
	}

	data, err := mpqHandle.ReadFile("data/global/excel/ItemStatCost.txt")
	if err != nil {
		log.Fatalf("opening file: %v", err)
	}

	fmt.Printf("length of file: %v bytes", len(data))
}
