package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"time"
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

	var repo *git.Repository
	if needToInit {
		repo, err = git.PlainInit(*output, false)
		if err != nil {
			fmt.Println("Failed to init git repository")
			log.Fatal(err)
		}
	} else {
		repo, err = git.PlainOpen(*output)
		if err != nil {
			fmt.Println("Failed to open existing git repository")
			log.Fatal(err)
		}
	}
	w, err := repo.Worktree()
	if err != nil {
		fmt.Println("Failed to open existing git repository")
		log.Fatal(err)
	}

	for _, file := range zipReader.File {
		if file.Name == "source/notes.json" {
			processNotes(output, w, file)
		}
	}
}

func processActiveNote(output *string, w *git.Worktree) func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	return func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		content, _ := jsonparser.GetString(value, "content")
		creationDate, _ := jsonparser.GetString(value, "creationDate")
		lastModified, _ := jsonparser.GetString(value, "lastModified")
		filename := writeFile(&content, output, &creationDate, &lastModified)
		if len(filename) > 0 {
			commitFile(w, &filename, "Exported "+filename)
		}
	}
}

func processTrashedNote(folder *string, w *git.Worktree) func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	return func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		content, _ := jsonparser.GetString(value, "content")
		creationDate, _ := jsonparser.GetString(value, "creationDate")
		lastModified, _ := jsonparser.GetString(value, "lastModified")
		filename := writeFile(&content, folder, &creationDate, &lastModified)
		if len(filename) > 0 {
			commitFile(w, &filename, "Exported to delete "+filename)
			err2 := os.Remove(path.Join(*folder, filename))
			if err2 != nil {
				fmt.Println("Unable to remove file from worktree: ", filename)
				log.Fatal(err2)
			}
			commitFile(w, &filename, "Deleted "+filename)
		}
	}
}

func commitFile(w *git.Worktree, filename *string, commitMessage string) {
	_, err := w.Add(*filename)
	if err != nil {
		fmt.Println("Unable to add file to worktree: ", filename)
		log.Fatal(err)
	}
	_, err2 := w.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "GitJournal Exporter",
			Email: "a@b.c",
			When:  time.Now(),
		},
	})
	if err2 != nil {
		fmt.Println("Unable to commit file:", filename)
		log.Fatal(err)
	}
}

func writeFile(content *string, folder *string, creationDate *string, lastModified *string) string {
	extractedName := extractFileName(content)
	if len(extractedName) == 0 {
		return ""
	}
	filename := extractedName + ".md"
	_, err := os.Stat(path.Join(*folder, filename))
	if !os.IsNotExist(err) {
		filename = extractedName + "_" + randomSuffix(5) + ".md"
	}
	file, err := os.Create(path.Join(*folder, filename))
	if err != nil {
		fmt.Println("Unable to create new file:", filename)
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprintln(file, "---")
	fmt.Fprintln(file, "created:", *creationDate)
	fmt.Fprintln(file, "modified:", *lastModified)
	fmt.Fprintln(file, "---")
	fmt.Fprint(file, *content)
	return filename
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomSuffix(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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

func processNotes(output *string, w *git.Worktree, file *zip.File) {
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
	_, err = jsonparser.ArrayEach(json, processActiveNote(output, w), "activeNotes")
	if err != nil {
		log.Fatal(err)
	}
	_, err = jsonparser.ArrayEach(json, processTrashedNote(output, w), "trashedNotes")
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
