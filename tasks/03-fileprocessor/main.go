package main

import (
	"log"

	fileprocessor "fileprocessor/fileprocessor"
)

func main() {
	lines, err := fileprocessor.ProcessFiles(fileprocessor.FS, "./src/file1.txt", "./src/file2.txt")
	if err != nil {
		log.Fatal("got error", err)
	}
	log.Println("total lines", lines)
}
