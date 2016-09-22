package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type RawComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	Email    string `json:"email"`
	URL      string `json:"www"`
	IP       string `json:"ip"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}

type ParsedComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	MD5      string `json:"md5"`
	URL      string `json:"www"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}

func BuildCommand() *cobra.Command {
	buildCmd := cobra.Command{
		Use:   "build",
		Short: "build",
		Run:   BuildComments,
	}

	return &buildCmd
}

func ParseComment(data *RawComment) *ParsedComment {
	return &ParsedComment{
		ID:       data.ID,
		ParentID: data.ParentID,
		Author:   data.Author,
		URL:      data.URL,
		Body:     data.Body,
		Date:     data.Date,
		MD5:      fmt.Sprintf("%x", md5.Sum([]byte(data.Email))),
	}
}

func GenerateComments(thread string) {
	comments, err := ioutil.ReadDir(path.Join("threads", thread))
	if err != nil {
		log.Fatalf("Failed to read thread %v: %v", thread, err)
	}

	output := []*ParsedComment{}
	for _, comment := range comments {
		if strings.HasSuffix(comment.Name(), ".json") {
			filePath := path.Join("threads", thread, comment.Name())
			reader, err := os.Open(filePath)
			if err != nil {
				log.Fatalf("Failed to open comment %v: %v", filePath, err)
			}
			defer reader.Close()

			data := &RawComment{}
			decoder := json.NewDecoder(reader)
			if err := decoder.Decode(data); err != nil {
				log.Fatalf("Failed to decode json for %v: %v", filePath, err)
			}
			output = append(output, ParseComment(data))
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

	countPath := path.Join("dist", thread+".count.json")
	count, err := os.Create(countPath)
	if err != nil {
		log.Fatalf("Error opening the count file %v: %v", countPath, err)
	}
	defer count.Close()
	count.WriteString(fmt.Sprintf("{\"count\": %v}", len(comments)))
}

func BuildComments(cmd *cobra.Command, args []string) {
	if err := os.MkdirAll("dist", 0777); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}

	threads, err := ioutil.ReadDir("threads")
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
