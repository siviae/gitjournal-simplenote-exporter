package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"log"
	"os"
)

func doMain(input *string, output *string) {
	zipReader, err := zip.OpenReader(*input)
	if err != nil {
		fmt.Println("Unable to open ", *input)
		log.Fatal(err)
	}
	defer zipReader.Close()

	err = os.Mkdir(*output, os.ModePerm)
	if err != nil {
		fmt.Println("Unable to create existing directory", *output)
	}

	for _, file := range zipReader.File {
		if file.Name == "source/notes.json" {
			processNotes(output, file)
		}
	}
}

func readActiveNote(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	fmt.Println(jsonparser.GetString(value, "content"))
}

func readTrashedNote(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	content, err := jsonparser.GetString(value, "content")
	fmt.Println("Trashed: ", content)
}

func processNotes(output *string, file *zip.File) {
	data, err := file.Open()
	if err != nil {
		fmt.Println("Unable to read source/notes.json from archive")
		log.Fatal(err)
	}
	defer data.Close()
	json, err := io.ReadAll(data)
	if err != nil {
		fmt.Println("Unable to read source/notes.json from archive")
		log.Fatal(err)
	}
	_, err = jsonparser.ArrayEach(json, readActiveNote, "activeNotes")
	if err != nil {
		log.Fatal(err)
	}
	_, err = jsonparser.ArrayEach(json, readTrashedNote, "trashedNotes")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	input := flag.String("input", "notes.zip", "Archive file, exported from simplenote.com")
	output := flag.String("output", ".", "Output folder (GitJournal repo)")
	flag.Parse()
	doMain(input, output)
}
