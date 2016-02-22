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

// ImportNodes reads data from nodes.dmp and write to bolt database
func ImportNodes(dbFile string, bucket string, dataFile string, batchSize int, force bool) {
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
		if len(items) != 13 {
			return nil, false, nil
		}
		return nodes.NodeFromArgs(items), true, nil
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
			node := data.(nodes.Node)
			nodeJSONStr, err := node.ToJSON()
			if err != nil {
				checkError(chunk.Err)
				return
			}
			records[i] = []string{node.TaxID, nodeJSONStr}
		}
		write2db(records, db, bucket)
		n += len(records)
		log.Info("%d records imported to %s", n, dbFile)
	}
}

var reDigitals = regexp.MustCompile(`^\d+$`)

// QueryNodeByTaxID querys Node by taxid
func QueryNodeByTaxID(db *bolt.DB, bucket string, taxids []string) ([]nodes.Node, error) {
	for _, taxid := range taxids {
		if !reDigitals.MatchString(taxid) {
			return []nodes.Node{}, fmt.Errorf("non-digital taxid given: %s", taxid)
		}
	}
	nods := make([]nodes.Node, len(taxids))
	if len(taxids) == 0 {
		return nods, nil
	}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("database not exists: %s", bucket)
		}
		for i, taxid := range taxids {
			s := string(b.Get([]byte(taxid)))
			if s == "" {
				nods[i] = nodes.Node{}
				continue
			}
			node, err := nodes.NodeFromJSON(s)
			if err != nil {
				return errors.New("failed to parse node record from database")
			}
			nods[i] = node
		}
		return nil
	})
	return nods, err
}

// LoadAllNodes loads all nodes into memory
func LoadAllNodes(db *bolt.DB, bucket string) (map[string]nodes.Node, error) {
	nods := make(map[string]nodes.Node)

	ch := make(chan string, runtime.NumCPU())
	chDone := make(chan int)
	go func() {
		for s := range ch {
			node, err := nodes.NodeFromJSON(s)
			checkError(err)
			nods[node.TaxID] = node
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
	return nods, err
}
