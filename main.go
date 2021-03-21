package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"log"
	"os"
	"path"
	"strings"
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

func processActiveNote(output *string) func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	return func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		content, err := jsonparser.GetString(value, "content")
		creationDate, err := jsonparser.GetString(value, "creationDate")
		lastModified, err := jsonparser.GetString(value, "lastModified")
		writeFile(output, &content, &creationDate, &lastModified)
	}
}

func writeFile(folder *string, content *string, creationDate *string, lastModified *string) {
	filename, done := extractFileName(content)
	if done {
		return
	}

	file, err := os.Create(path.Join(*folder, filename+".md"))
	if err != nil {
		fmt.Println("Unable to read source/notes.json from archive")
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprintln(file, "---")
	fmt.Fprintln(file, "created: ", creationDate)
	fmt.Fprintln(file, "modified: ", lastModified)
	fmt.Fprintln(file, "---")
	fmt.Fprint(file, content)
}

func extractFileName(content *string) (string, bool) {
	newLinePos := strings.Index(*content, "\n")
	if newLinePos == -1 {
		newLinePos = len(*content)
	}
	if newLinePos > 64 {
		newLinePos = 64
	}
	if newLinePos == 0 {
		return "", true
	}
	filename := strings.Replace((*content)[:newLinePos], string(os.PathSeparator), "", -1)
	return filename, false
}

func processTrashedNote(output *string) func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	return func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		content, err := jsonparser.GetString(value, "content")
		fmt.Println("Trashed: ", content)
	}
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
	_, err = jsonparser.ArrayEach(json, processActiveNote(output), "activeNotes")
	if err != nil {
		log.Fatal(err)
	}
	_, err = jsonparser.ArrayEach(json, processTrashedNote(output), "trashedNotes")
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
