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
	"os"

	"github.com/boltdb/bolt"
	"github.com/mitchellh/go-homedir"
	"github.com/shenwei356/util/pathutil"
)

// DbPath is the database path
var DbPath string
var defaultDbPath string

// DbFile is the bolt database filename
var DbFile string
var defaultDbFile string

func init() {
	var err error
	defaultDbPath, err = homedir.Expand("~/.gotaxon/")
	if err != nil {
		log.Error(err)
	}
	defaultDbFile = "db.db"

}

// InitDatabase initializes database direcotry
func InitDatabase(dbPath string, dbFile string, force bool) {
	if dbPath == "" {
		DbPath = defaultDbPath
	} else {
		DbPath = dbPath
	}
	if dbFile == "" {
		DbFile = defaultDbFile
	} else {
		DbFile = dbFile
	}

	existed, err := pathutil.DirExists(DbPath)
	checkError(err)
	if existed {
		if force {
			os.RemoveAll(DbPath)
			log.Info("Remove old dbPath and recreate: %s", DbPath)
			os.MkdirAll(DbPath, os.ModePerm)
		} else {
			log.Info("Database directory (%s) existed. Nothing happended. Use --force to reinit", DbPath)
			return
		}
	} else {
		os.MkdirAll(DbPath, os.ModePerm)
	}
}

var pool *DBPool

// DBPool is bolt db connection pool
type DBPool struct {
	dbs []*bolt.DB
	ch  chan *bolt.DB
}

// NewDBPool is constructor for DBPools
func NewDBPool(dbFilePath string, n int) *DBPool {
	pool := new(DBPool)
	pool.dbs = make([]*bolt.DB, n)
	pool.ch = make(chan *bolt.DB, n)

	for i := 0; i < n; i++ {
		db, err := bolt.Open(dbFilePath, 0600, &bolt.Options{ReadOnly: true})
		checkError(err)
		pool.dbs[i] = db
		pool.ch <- db
	}

	return pool
}

// GetDB gets one connection
func (p *DBPool) GetDB() *bolt.DB {
	return <-p.ch
}

// ReleaseDB releases a connection
func (p *DBPool) ReleaseDB(db *bolt.DB) {
	p.ch <- db
}

// Close closes all connection
func (p *DBPool) Close() {
	for _, db := range p.dbs {
		db.Close()
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

func write2db(kvs [][]string, db *bolt.DB, bucket string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}
		for _, items := range kvs {
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
