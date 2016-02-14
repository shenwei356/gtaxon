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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

// StartServer runs a web server for query
func StartServer(dbFilePath string, port int, timeout int, threads int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	pool = NewDBPool(dbFilePath, threads)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/gi2taxid", gi2taxid)

	// router.Run(fmt.Sprintf(":%d", port))
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

// ErrDBNotFound is a error
var ErrDBNotFound = errors.New("DB not found")

type gi2Taxid struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	Gis    []string `json:"gis"`
	Taxids []string `json:"taxids"`
}

func gi2taxid(c *gin.Context) {
	var msg gi2Taxid

	// gi := c.Query("gi")  // single value
	// multiple values
	c.Request.ParseForm()
	gis := c.Request.Form["gi"]

	if gis == nil {
		msg.Status = "FAILED"
		msg.Message = "no GIs given"
		c.JSON(http.StatusOK, msg)
		return
	}

	bucket := c.DefaultQuery("db", "gi_taxid_prot")
	if bucket != "gi_taxid_prot" && bucket != "gi_taxid_nucl" {
		msg.Status = "FAILED"
		msg.Message = fmt.Sprintf("invalid db: %s. valid: gi_taxid_prot or gi_taxid_nucl", bucket)
		c.JSON(http.StatusOK, msg)
		return
	}

	db := pool.GetDB()
	defer pool.ReleaseDB(db)

	taxids := make([]string, len(gis))
	n := 0 // counter of seccessful query
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrDBNotFound
		}
		for i, gi := range gis {
			taxid := string(b.Get([]byte(gi)))
			if taxid != "" {
				n++
			}
			taxids[i] = taxid
		}
		return nil
	})
	if err != nil {
		if err == ErrDBNotFound {
			msg.Status = "FAILED"
			msg.Message = fmt.Sprintf("DB not found: %s", bucket)
			c.JSON(http.StatusOK, msg)
			return
		}
		msg.Status = "FAILED"
		msg.Message = fmt.Sprintf("error: %s", err)
		c.JSON(http.StatusOK, msg)
		return
	}
	msg.Status = "OK"
	msg.Message = fmt.Sprintf("sum: %d, found: %d", len(gis), n)
	msg.Gis = gis
	msg.Taxids = taxids
	c.JSON(http.StatusOK, msg)
}

// RemoteQueryGi2Taxid query from remote server
func RemoteQueryGi2Taxid(host string, port int, dbType string, gis []string) ([]string, []string) {
	host = strings.TrimSpace(host)
	var url string
	if regexp.MustCompile("^http://").MatchString(host) {
		url = fmt.Sprintf("%s:%d/gi2taxid", host, port)
	} else {
		url = fmt.Sprintf("http://%s:%d/gi2taxid", host, port)
	}

	request := gorequest.New().Get(url).Set("db", dbType)

	for _, gi := range gis {
		request = request.Param("gi", gi)
	}

	_, body, errs := request.End()
	if errs != nil {
		log.Error(errs)
		os.Exit(-1)
	}

	var result gi2Taxid
	err := json.Unmarshal([]byte(body), &result)
	checkError(err)

	return result.Gis, result.Taxids
}
