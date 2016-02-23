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
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/gtaxon/taxon/nodes"
)

// ImportDivisions reads data from divisions.dmp and write to bolt database
func ImportDivisions(dbFile string, bucket string, dataFile string, batchSize int, force bool) {
	db, err := bolt.Open(dbFile, 0600, nil)
	checkError(err)
	defer db.Close()

	if force {
		err = deleteBucket(db, bucket)
		checkError(err)
		log.Info("Old database deleted: %s", bucket)
	}

	if batchSize <= 0 {
		batchSize = 10000
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
		return nodes.DivisionFromArgs(items), true, nil
	}

	reader, err := breader.NewBufferedReader(dataFile, runtime.NumCPU(), batchSize, fn)
	checkError(err)

	n := 0
	for chunk := range reader.Ch {
		if chunk.Err != nil {
			checkError(chunk.Err)
			return
		}

		records := make([][]string, len(chunk.Data))
		for i, data := range chunk.Data {
			division := data.(nodes.Division)
			divisionJSONStr, err := division.ToJSON()
			if err != nil {
				checkError(chunk.Err)
				return
			}
			records[i] = []string{division.DivisionID, divisionJSONStr}
		}
		write2db(records, db, bucket)
		n += len(records)
		log.Info("%d records imported to %s", n, dbFile)
	}
}

// QueryDivisionByDivisionID querys Division by taxid
func QueryDivisionByDivisionID(db *bolt.DB, bucket string, ids []string) ([]nodes.Division, error) {
	for _, id := range ids {
		if !reDigitals.MatchString(id) {
			return []nodes.Division{}, fmt.Errorf("non-digital division given: %s", id)
		}
	}
	divisions := make([]nodes.Division, len(ids))
	if len(ids) == 0 {
		return divisions, nil
	}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("database not exists: %s", bucket)
		}
		for i, id := range ids {
			s := string(b.Get([]byte(id)))
			if s == "" {
				divisions[i] = nodes.Division{}
				continue
			}
			division, err := nodes.DivisionFromJSON(s)
			if err != nil {
				return errors.New("failed to parse division record from database")
			}
			divisions[i] = division
		}
		return nil
	})
	return divisions, err
}

// LoadAllDivisions loads all divisions into memory
func LoadAllDivisions(db *bolt.DB, bucket string) (map[string]nodes.Division, error) {
	divisions := make(map[string]nodes.Division)

	ch := make(chan string, runtime.NumCPU())
	chDone := make(chan int)
	go func() {
		for s := range ch {
			division, err := nodes.DivisionFromJSON(s)
			checkError(err)
			divisions[division.DivisionID] = division
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
	return divisions, err
}
