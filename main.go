package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/go-git/go-git/v5"
	"io"
	"log"
	"os"
	"path"
)

func doMain(input *string, output *string) {
	zipReader, err := zip.OpenReader(*input)
	if err != nil {
		fmt.Println("Unable to open ", *input)
		log.Fatal(err)
	}
	defer zipReader.Close()

	err = os.Mkdir(*output, os.ModePerm)
	needToInit := true
	if err != nil {
		fmt.Println("Unable to create existing directory, expecting a git repository", *output)
		needToInit = false
	}

	if needToInit {
		_, err := git.PlainInit(*output, false)
		if err != nil {
			fmt.Println("Failed to init git repository")
			log.Fatal(err)
		}
	} else {
		_, err := git.PlainOpen(*output)
		if err != nil {
			fmt.Println("Failed to open existing git repository")
			log.Fatal(err)
		}
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
	filename := extractFileName(content)
	if len(filename) == 0 {
		return
	}

	file, err := os.Create(path.Join(*folder, filename+".md"))
	if err != nil {
		fmt.Println("Unable to read source/notes.json from archive")
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprintln(file, "---")
	fmt.Fprintln(file, "created: ", *creationDate)
	fmt.Fprintln(file, "modified: ", *lastModified)
	fmt.Fprintln(file, "---")
	fmt.Fprint(file, *content)
}

func extractFileName(content *string) string {
	size := 0
	filename := make([]rune, 0, 128)
	for _, char := range *content {
		if size >= 64 {
			break
		}
		if char == '\n' {
			break
		}
		if char == os.PathSeparator || char == '\r' {
			continue
		}
		filename = append(filename, char)
		size++
	}
	return string(filename)
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
