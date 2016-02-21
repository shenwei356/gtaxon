// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package taxon

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/gtaxon/taxon/nodes"
)

// ImportNames reads
func ImportNames(dbFile string, bucket string, dataFile string, chunkSize int, force bool) {
	db, err := bolt.Open(dbFile, 0600, nil)
	checkError(err)
	defer db.Close()

	if force {
		err = deleteBucket(db, bucket)
		checkError(err)
		log.Info("Old database deleted: %s", bucket)
	}

	if chunkSize <= 0 {
		chunkSize = 10000
	}

	re := regexp.MustCompile(`\t\|$`)
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimRight(line, "\n")
		if line == "" {
			return nil, false, nil
		}
		items := strings.Split(re.ReplaceAllString(line, ""), "\t|\t")
		if len(items) != 4 {
			return nil, false, nil
		}
		return nodes.NameFromArgs(items), true, nil
	}

	reader, err := breader.NewBufferedReader(dataFile, runtime.NumCPU(), chunkSize, fn)
	checkError(err)

	names := make(map[string]nodes.Name)
	n := 0
	for chunk := range reader.Ch {
		if chunk.Err != nil {
			checkError(chunk.Err)
			return
		}

		for _, data := range chunk.Data {
			name := data.(nodes.Name)
			if _, ok := names[name.TaxID]; ok {
				names[name.TaxID] = nodes.MergeNames(names[name.TaxID], name)
			} else {
				names[name.TaxID] = name
			}
		}
		n += len(chunk.Data)
		log.Info("%d records readed", n)
	}

	chResults := make(chan []string, runtime.NumCPU())

	// write to db
	chDone := make(chan int)
	go func() {
		records := make([][]string, chunkSize)
		i := 0
		n := 0
		for s := range chResults {
			records[i] = s
			i++
			n++
			if i%chunkSize == 0 {
				write2db(records, db, bucket)
				log.Info("%d records imported to %s", n, dbFile)
				records = make([][]string, chunkSize)
				i = 0
			}
		}
		log.Info("%d records imported to %s", n, dbFile)
		chDone <- 1
	}()

	// name to json
	tokens := make(chan int, runtime.NumCPU())
	var wg sync.WaitGroup
	for _, name := range names {
		tokens <- 1
		wg.Add(1)
		go func(name nodes.Name) {
			defer func() {
				wg.Done()
				<-tokens
			}()
			nameJSONStr, err := name.ToJSON()
			checkError(err)
			chResults <- []string{name.TaxID, nameJSONStr}
		}(name)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

// LoadAllNames loads all names into memory
func LoadAllNames(db *bolt.DB, bucket string) (map[string]nodes.Name, error) {
	names := make(map[string]nodes.Name)

	ch := make(chan string, runtime.NumCPU())
	chDone := make(chan int)
	go func() {
		for s := range ch {
			name, err := nodes.NameFromJSON(s)
			checkError(err)
			names[name.TaxID] = name
		}
		chDone <- 1
	}()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("database not exists: %s", bucket)
		}

		b.ForEach(func(k, v []byte) error {
			ch <- string(v)
			return nil
		})

		return nil
	})
	close(ch)
	<-chDone
	return names, err
}

// QueryTaxIDByName query taxid by name
func QueryTaxIDByName(db *bolt.DB, bucket string, useRegexp bool, nameClass string, queries []string) (map[string][]string, error) {
	if nodes.Names == nil {
		log.Info("load all names ...")
		names, err := LoadAllNames(db, bucket)
		if err != nil {
			return nil, err
		}
		nodes.Names = names
		log.Info("load all names ... done")
	}

	var queryRegexps map[string]*regexp.Regexp
	if useRegexp {
		queryRegexps = make(map[string]*regexp.Regexp, len(queries))
		for _, query := range queries {
			queryRegexps[query] = regexp.MustCompile(query)
		}
	}

	result := make(map[string][]string)

	chResult := make(chan []string, runtime.NumCPU())
	chDone := make(chan int)
	go func() {
		for s := range chResult {
			if _, ok := result[s[0]]; ok {
				result[s[0]] = append(result[s[0]], s[1])
			} else {
				result[s[0]] = []string{s[1]}
			}
		}
		chDone <- 1
	}()

	var wg sync.WaitGroup
	tokens := make(chan int, runtime.NumCPU())
	var limitNameClass bool
	if nameClass != "" {
		limitNameClass = true
	}
	for _, name := range nodes.Names {
		wg.Add(1)
		tokens <- 1
		go func(name nodes.Name) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			for _, query := range queries {
				for _, nameItem := range name.Names {
					if limitNameClass && nameItem.NameClass != nameClass {
						continue
					}
					match := false
					if useRegexp {
						if queryRegexps[query].MatchString(nameItem.Name) {
							match = true
						}
					} else {
						if query == nameItem.Name {
							match = true
						}
					}

					if match {
						chResult <- []string{query, name.TaxID}
						break
					}
				}
			}
		}(name)
	}
	wg.Wait()
	close(chResult)
	<-chDone

	return result, nil
}
