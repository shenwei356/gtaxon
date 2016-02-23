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
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/shenwei356/gtaxon/taxon/nodes"
)

var threadNum int

// StartServer runs a web server for query
func StartServer(dbFilePath string, port int, timeout int, threads int) {
	runtime.GOMAXPROCS(threads)
	threadNum = threads
	pool = NewDBPool(dbFilePath, threads)

	done := make(chan int)
	go func() {
		db := pool.GetDB()
		defer pool.ReleaseDB(db)

		log.Info("load all names ...")
		names, err := LoadAllNames(db, "names")
		checkError(err)
		nodes.SetNames(names)
		log.Info("load all names ... done")
		done <- 1
	}()

	done1 := make(chan int)
	go func() {
		db := pool.GetDB()
		defer pool.ReleaseDB(db)

		log.Info("load all nodes ...")
		nods, err := LoadAllNodes(db, "nodes")
		nodes.SetNodes(nods)
		checkError(err)
		log.Info("load all nodes ... done")

		done1 <- 1
	}()

	done2 := make(chan int)
	go func() {
		db := pool.GetDB()
		defer pool.ReleaseDB(db)

		log.Info("load all divisions ...")
		divisions, err := LoadAllDivisions(db, "divisions")
		nodes.SetDivisions(divisions)
		checkError(err)
		log.Info("load all divisions ... done")

		done2 <- 1
	}()

	done3 := make(chan int)
	go func() {
		db := pool.GetDB()
		defer pool.ReleaseDB(db)

		log.Info("load all gencodes ...")
		gencodes, err := LoadAllGenCodes(db, "gencodes")
		nodes.SetGenCodes(gencodes)
		checkError(err)
		log.Info("load all gencodes ... done")

		done3 <- 1
	}()

	<-done2
	<-done3
	<-done
	<-done1

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/gi2taxid", gi2taxid)
	router.GET("/taxid2taxon", taxid2taxon)
	router.GET("/name2taxid", name2taxid)
	router.GET("/lca", lca)

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

// --------------------------------------------------------------------------

// MessageLCAMap is
type MessageLCAMap struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	LCA map[string]nodes.Taxon `json:"taxids2taxon"`
}

func lca(c *gin.Context) {
	var msg MessageLCAMap

	c.Request.ParseForm()
	queries := c.Request.Form["taxids"]

	if queries == nil {
		msg.Status = "FAILED"
		msg.Message = "no Taxids given"
		c.JSON(http.StatusOK, msg)
		return
	}

	db := pool.GetDB()
	defer pool.ReleaseDB(db)

	result := make(map[string]nodes.Node)
	for _, query := range queries {
		taxids := strings.Split(query, ",")
		lca, err := nodes.LCA(nodes.Nodes, taxids)
		if err != nil {
			msg.Status = "FAILED"
			msg.Message = fmt.Sprintf("error: %s", err)
			c.JSON(http.StatusOK, msg)
			return
		}
		result[query] = lca
	}

	msg.Status = "OK"
	msg.Message = fmt.Sprintf("sum: %d", len(queries))
	msg.LCA = make(map[string]nodes.Taxon, len(result))
	for k, node := range result {
		msg.LCA[k], _ = nodes.GetTaxonByTaxID(node.TaxID)
	}

	c.JSON(http.StatusOK, msg)
}

// RemoteQueryLCA is
func RemoteQueryLCA(host string, port int, queries []string) MessageLCAMap {
	host = strings.TrimSpace(host)
	var url string
	if regexp.MustCompile("^http://").MatchString(host) {
		url = fmt.Sprintf("%s:%d/lca", host, port)
	} else {
		url = fmt.Sprintf("http://%s:%d/lca", host, port)
	}

	request := gorequest.New().Get(url)

	for _, query := range queries {
		request = request.Param("taxids", query)
	}

	_, body, errs := request.End()
	if errs != nil {
		log.Error(errs)
		os.Exit(-1)
	}

	var result MessageLCAMap
	err := json.Unmarshal([]byte(body), &result)
	checkError(err)

	return result
}

// --------------------------------------------------------------------------

// MessageTaxid2TaxonMap is
type MessageTaxid2TaxonMap struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	Taxons map[string]nodes.Taxon `json:"taxid2taxon"`
}

func taxid2taxon(c *gin.Context) {
	var msg MessageTaxid2TaxonMap

	c.Request.ParseForm()
	taxids := c.Request.Form["taxid"]

	if taxids == nil {
		msg.Status = "FAILED"
		msg.Message = "no Taxids given"
		c.JSON(http.StatusOK, msg)
		return
	}

	db := pool.GetDB()
	defer pool.ReleaseDB(db)

	bucket := "nodes"
	nods, err := QueryNodeByTaxID(db, bucket, taxids)
	if err != nil {
		msg.Status = "FAILED"
		msg.Message = fmt.Sprintf("error: %s", err)
		c.JSON(http.StatusOK, msg)
		return
	}
	msg.Status = "OK"
	msg.Message = fmt.Sprintf("sum: %d", len(nods))
	msg.Taxons = make(map[string]nodes.Taxon)
	for _, node := range nods {
		taxon, _ := nodes.GetTaxonByTaxID(node.TaxID)
		msg.Taxons[node.TaxID] = taxon
	}
	c.JSON(http.StatusOK, msg)
}

// RemoteQueryTaxid2Taxon is
func RemoteQueryTaxid2Taxon(host string, port int, taxids []string) MessageTaxid2TaxonMap {
	host = strings.TrimSpace(host)
	var url string
	if regexp.MustCompile("^http://").MatchString(host) {
		url = fmt.Sprintf("%s:%d/taxid2taxon", host, port)
	} else {
		url = fmt.Sprintf("http://%s:%d/taxid2taxon", host, port)
	}

	request := gorequest.New().Get(url)

	for _, taxid := range taxids {
		request = request.Param("taxid", taxid)
	}

	_, body, errs := request.End()
	if errs != nil {
		log.Error(errs)
		os.Exit(-1)
	}

	var result MessageTaxid2TaxonMap
	err := json.Unmarshal([]byte(body), &result)
	checkError(err)

	return result
}

// --------------------------------------------------------------------------

// MssageName2TaxIDMap is
type MssageName2TaxIDMap struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	TaxIDs map[string][]TaxIDSciNameItem `json:"name2taxid"`
}

// TaxIDSciNameItem is
type TaxIDSciNameItem struct {
	TaxID          int
	ScientificName string
}

func name2taxid(c *gin.Context) {
	var msg MssageName2TaxIDMap

	c.Request.ParseForm()
	useRegexp := false
	if c.Query("regexp") != "" {
		useRegexp = true
	}
	nameClass := ""
	if c.Query("class") != "" {
		nameClass = c.Query("class")
	}
	names := c.Request.Form["name"]

	if names == nil {
		msg.Status = "FAILED"
		msg.Message = "no names given"
		c.JSON(http.StatusOK, msg)
		return
	}

	db := pool.GetDB()
	defer pool.ReleaseDB(db)

	bucket := "names"
	results, err := QueryTaxIDByName(db, bucket, useRegexp, nameClass, threadNum*2, names)
	if err != nil {
		msg.Status = "FAILED"
		msg.Message = fmt.Sprintf("error: %s", err)
		c.JSON(http.StatusOK, msg)
		return
	}
	msg.Status = "OK"
	msg.Message = fmt.Sprintf("sum: %d", len(names))

	name2taxidResults := make(map[string][]TaxIDSciNameItem, len(results))
	for name, taxids := range results {
		taxidsSciNameItems := make([]TaxIDSciNameItem, len(taxids))
		for i, taxid := range taxids {
			node := nodes.Nodes[taxid]
			taxidInt, _ := strconv.Atoi(node.TaxID)
			scientificName := ""
			for _, nameItem := range nodes.Names[node.TaxID].Names {
				if nameItem.NameClass == "scientific name" {
					scientificName = nameItem.Name
				}
			}
			taxidsSciNameItems[i] = TaxIDSciNameItem{
				TaxID:          taxidInt,
				ScientificName: scientificName,
			}
		}
		name2taxidResults[name] = taxidsSciNameItems
	}
	msg.TaxIDs = name2taxidResults
	c.JSON(http.StatusOK, msg)
}

// RemoteQueryName2TaxID is
func RemoteQueryName2TaxID(host string, port int, useRegexp bool, nameClass string, names []string) MssageName2TaxIDMap {
	host = strings.TrimSpace(host)
	var url string
	if regexp.MustCompile("^http://").MatchString(host) {
		url = fmt.Sprintf("%s:%d/name2taxid", host, port)
	} else {
		url = fmt.Sprintf("http://%s:%d/name2taxid", host, port)
	}

	request := gorequest.New().Get(url)
	if useRegexp {
		request = request.Param("regexp", "1")
	}
	if nameClass != "" {
		request = request.Param("class", nameClass)
	}

	for _, name := range names {
		request = request.Param("name", name)
	}

	_, body, errs := request.End()
	if errs != nil {
		log.Error(errs)
		os.Exit(-1)
	}

	var result MssageName2TaxIDMap
	err := json.Unmarshal([]byte(body), &result)
	checkError(err)
	return result
}

// --------------------------------------------------------------------------

// MessageGI2TaxidMap is
type MessageGI2TaxidMap struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	Taxids map[string]string `json:"gi2taxid"`
}

func gi2taxid(c *gin.Context) {
	var msg MessageGI2TaxidMap

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

	taxids := make(map[string]string, len(gis))
	n := 0 // counter of seccessful query
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("db not found: %s", bucket)
		}
		for _, gi := range gis {
			taxid := string(b.Get([]byte(gi)))
			if taxid != "" {
				n++
			}
			taxids[gi] = taxid
		}
		return nil
	})
	if err != nil {
		msg.Status = "FAILED"
		msg.Message = fmt.Sprintf("error: %s", err)
		c.JSON(http.StatusOK, msg)
		return
	}
	msg.Status = "OK"
	msg.Message = fmt.Sprintf("sum: %d, found: %d", len(gis), n)
	msg.Taxids = taxids
	c.JSON(http.StatusOK, msg)
}

// RemoteQueryGi2Taxid query from remote server
func RemoteQueryGi2Taxid(host string, port int, dbType string, gis []string) MessageGI2TaxidMap {
	host = strings.TrimSpace(host)
	var url string
	if regexp.MustCompile("^http://").MatchString(host) {
		url = fmt.Sprintf("%s:%d/gi2taxid", host, port)
	} else {
		url = fmt.Sprintf("http://%s:%d/gi2taxid", host, port)
	}

	request := gorequest.New().Get(url).Param("db", dbType)

	for _, gi := range gis {
		request = request.Param("gi", gi)
	}

	_, body, errs := request.End()
	if errs != nil {
		log.Error(errs)
		os.Exit(-1)
	}

	var result MessageGI2TaxidMap
	err := json.Unmarshal([]byte(body), &result)
	checkError(err)

	return result
}
