package comments

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/netlify/gotell/conf"
)

var (
	filenameRegexp = regexp.MustCompile(`(?:[^/]+)/(\d+)/(\d+)/([^/]+)`)
)

func Build(config *conf.Configuration) {
	if err := os.MkdirAll(config.Threads.Destination, 0755); err != nil {
		logrus.Fatalf("Failed to create output dir: %v", err)
	}

	threads, err := filepath.Glob(config.Threads.Source + "/*/*/*")
	if err != nil {
		logrus.Fatalf("Failed to list threads: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan int, 100)

	for _, thread := range threads {
		sem <- 1
		wg.Add(1)
		go func(t string) {
			generate(t, config.Threads.Destination)
			<-sem
			wg.Done()
		}(thread)
	}

	wg.Wait()

	emptyPath := path.Join(config.Threads.Destination, "empty.json")
	empty, err := os.Create(emptyPath)
	if err != nil {
		log.Fatalf("Error opening the empty file %v: %v", emptyPath, err)
	}
	defer empty.Close()
	empty.WriteString("[]")

	countPath := path.Join(config.Threads.Destination, "empty.count.json")
	count, err := os.Create(countPath)
	if err != nil {
		log.Fatalf("Error opening the count file %v: %v", countPath, err)
	}
	defer count.Close()
	count.WriteString(fmt.Sprintf("{\"count\": %v}", 0))

	redirectsPath := path.Join(config.Threads.Destination, "_redirects")
	redirects, err := os.Create(redirectsPath)
	if err != nil {
		log.Fatalf("Error opening the redirects file %v: %v", redirectsPath, err)
	}
	defer redirects.Close()
	redirects.WriteString(`
/*.count.json  /empty.count.json   200
/*.json  /empty.json  200
`)
}

func generate(source, dest string) {
	comments, err := ioutil.ReadDir(source)
	if err != nil {
		log.Fatalf("Failed to read thread %v: %v", source, err)
	}

	output := []*ParsedComment{}
	for _, comment := range comments {
		if strings.HasSuffix(comment.Name(), ".json") {
			filePath := path.Join(source, comment.Name())
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

	matches := filenameRegexp.FindStringSubmatch(source)
	name := matches[1] + "-" + matches[2] + "-" + matches[3]

	distPath := path.Join(dest, name+".json")
	dist, err := os.Create(distPath)
	if err != nil {
		log.Fatalf("Error opening output file %v: %v", distPath, err)
	}
	defer dist.Close()

	encoder := json.NewEncoder(dist)
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("Failed to encode json for %v: %v", distPath, err)
	}

	countPath := path.Join(dest, name+".count.json")
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
		Twitter:  data.Twitter,
		URL:      data.URL,
		Body:     data.Body,
		Date:     data.Date,
		MD5:      fmt.Sprintf("%x", md5.Sum([]byte(data.Email))),
	}
}
