package comments

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

	"github.com/Sirupsen/logrus"
	"github.com/netlify/gotell/conf"
)

func Build(config *conf.Configuration) {
	if err := os.MkdirAll(config.Threads.Destination, 0755); err != nil {
		logrus.Fatalf("Failed to create output dir: %v", err)
	}

	threads, err := ioutil.ReadDir(config.Threads.Source)
	if err != nil {
		logrus.Fatalf("Failed to list threads: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan int, 100)

	for _, info := range threads {
		if info.IsDir() {
			sem <- 1
			wg.Add(1)
			go func() {
				generate(config.Threads.Source, config.Threads.Destination, info.Name())
				<-sem
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

func generate(source, dest, thread string) {
	comments, err := ioutil.ReadDir(path.Join(source, thread))
	if err != nil {
		log.Fatalf("Failed to read thread %v: %v", thread, err)
	}

	output := []*ParsedComment{}
	for _, comment := range comments {
		if strings.HasSuffix(comment.Name(), ".json") {
			filePath := path.Join(source, thread, comment.Name())
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
			output = append(output, ParseRaw(data))
		}
	}

	distPath := path.Join(dest, thread+".json")
	dist, err := os.Create(distPath)
	if err != nil {
		log.Fatalf("Error opening output file %v: %v", distPath, err)
	}
	defer dist.Close()

	encoder := json.NewEncoder(dist)
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("Failed to encode json for %v: %v", distPath, err)
	}

	countPath := path.Join(dest, thread+".count.json")
	count, err := os.Create(countPath)
	if err != nil {
		log.Fatalf("Error opening the count file %v: %v", countPath, err)
	}
	defer count.Close()
	count.WriteString(fmt.Sprintf("{\"count\": %v}", len(comments)))
}

func ParseRaw(data *RawComment) *ParsedComment {
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
