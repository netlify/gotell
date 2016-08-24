package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type Comment struct {
	Author string `json:"author"`
	Email  string `json:"email"`
	Url    string `json:"www"`
	IP     string `json:"ip"`
	Body   string `json:"body"`
	Date   string `json:"date"`
}

func GenerateComments(thread string) {
	dir, err := os.Open(path.Join("threads", thread))
	if err != nil {
		log.Fatalf("Failed to open thread %v: %v", thread, err)
	}
	defer dir.Close()

	comments, err := dir.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to read thread %v: %v", thread, err)
	}

	output := []*Comment{}
	for _, comment := range comments {
		if strings.HasSuffix(comment.Name(), ".json") {
			filePath := path.Join("threads", thread, comment.Name())
			reader, err := os.Open(filePath)
			if err != nil {
				log.Fatalf("Failed to open comment %v: %v", filePath, err)
			}
			defer reader.Close()

			data := &Comment{}
			decoder := json.NewDecoder(reader)
			if err := decoder.Decode(data); err != nil {
				log.Fatalf("Failed to decode json for %v: %v", filePath, err)
			}
			output = append(output, data)
		}
	}

	distPath := path.Join("dist", thread+".json")
	dist, err := os.Create(distPath)
	if err != nil {
		log.Fatalf("Error opening output file %v: %v", distPath, err)
	}
	defer dist.Close()
	encoder := json.NewEncoder(dist)
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("Failed to encode json for %v: %v", distPath, err)
	}
}

func main() {
	if err := os.MkdirAll("dist", 0777); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}

	dir, err := os.Open("threads")
	if err != nil {
		log.Fatalf("Failed to open `threads` dir: %v", err)
	}
	defer dir.Close()
	threads, err := dir.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to list threads: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan int, 100)

	for _, info := range threads {
		if info.IsDir() {
			sem <- 1
			wg.Add(1)
			go func() {
				defer func() {
					<-sem
					wg.Done()
				}()
				GenerateComments(info.Name())
			}()
		}
	}
	wg.Wait()
}
