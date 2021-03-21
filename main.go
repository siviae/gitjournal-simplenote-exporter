package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"log"
)

func main() {
	input := flag.String("input", "notes.zip", "Archive file, exported from simplenote.com")
	output := flag.String("output", ".", "Output folder (GitJournal repo)")
	flag.Parse()

	zipReader, err := zip.OpenReader(*input)
	if err != nil {
		fmt.Println("Unable to open ", *input)
		log.Fatal(err)
	}
	fmt.Println("Output is: ", *output)
	for _, file := range zipReader.File {
		fmt.Println(file.Name)
	}

	err = zipReader.Close()
	if err != nil {
		log.Fatal(err)
	}
}
