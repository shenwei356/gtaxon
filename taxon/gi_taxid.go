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
	"runtime"
	"strings"

	"github.com/boltdb/bolt"
	fileutil "github.com/shenwei356/util/file"
)

// LoadGiTaxid reads gi_taxid_nucl or gi_taxid_prot file and writes the data to database
func LoadGiTaxid(dbFile string, bucket string, dataFile string, batchSize int, force bool) {
	db, err := bolt.Open(dbFile, 0600, nil)
	checkError(err)
	defer db.Close()

	if force {
		err = deleteBucket(db, bucket)
		checkError(err)
		log.Info("Old database deleted: %s", bucket)
	}

	if batchSize <= 0 {
		batchSize = 1000000
	}

	fn := func(line string) (string, bool) {
		if line == "" {
			return "", false
		}
		return line, true
	}
	chRead, err := fileutil.ReadFileWithBuffer(dataFile, batchSize, runtime.NumCPU(), fn)
	checkError(err)

	n := 0
	for batch := range chRead {
		var records [][]string

		for _, line := range batch {
			items := strings.Split(line, "\t")
			if len(items) != 2 {
				continue
			}
			records = append(records, items)
		}
		write2db(records, db, bucket)
		n += len(records)
		log.Info("%d records imported to %s", n, dbFile)
	}
}

func deleteBucket(db *bolt.DB, bucket string) error {
	// create the bucket if it not exists
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to delete bucket: %s", err)
		}
		return nil
	})
	return err
}

func write2db(buffer [][]string, db *bolt.DB, bucket string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}
		for _, items := range buffer {
			if len(items) == 0 {
				break
			}
			err = b.Put([]byte(items[0]), []byte(items[1]))
			if err != nil {
				return fmt.Errorf("failed to put record: %s:%s", items[0], items[1])
			}
		}
		return nil
	})
	return err
}

// QueryGi2Taxid querys taxids by gis
func QueryGi2Taxid(db *bolt.DB, bucket string, gis []string) ([]string, error) {
	taxids := make([]string, len(gis))
	if len(gis) == 0 {
		return taxids, nil
	}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("database not exists: %s", bucket)
		}
		for i, gi := range gis {
			taxids[i] = string(b.Get([]byte(gi)))
		}
		return nil
	})
	return taxids, err
}
