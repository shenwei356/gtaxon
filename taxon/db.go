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
	"os"

	"github.com/mitchellh/go-homedir"
	pathutil "github.com/shenwei356/util/path"
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
