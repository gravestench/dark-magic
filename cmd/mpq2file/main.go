package main

import (
	"fmt"
	"io"
	"log"

	"github.com/gravestench/mpq"
)

func main() {
	mpqHandle, err := mpq.New("/Users/dylanknuth/d2_english_mpq/patch_d2.mpq")
	if err != nil {
		log.Fatalf("opening mpq archive: %v", err)
	}

	fileHandle, err := mpqHandle.ReadFileStream("data/global/excel/ItemStatCost.txt")
	if err != nil {
		log.Fatalf("opening file: %v", err)
	}

	data, err := io.ReadAll(fileHandle)
	if err != nil {
		log.Fatalf("reading file data: %v", err)
	}

	fmt.Printf("length of file: %v bytes", len(data))
}
