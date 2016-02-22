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

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/gtaxon/taxon"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "query from local database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runtime.GOMAXPROCS(runtime.NumCPU())

		queryType, err := cmd.Flags().GetString("type")
		checkError(err)
		dataFile, err := cmd.Flags().GetString("file")
		checkError(err)
		chunkSize, err := cmd.Flags().GetInt("chunk-size")
		checkError(err)
		threads, err := cmd.Flags().GetInt("threads")
		checkError(err)

		if dataFile == "" {
			if len(args) == 0 {
				log.Error("Queries needed. Type \"gtaxon cli local -h\" for help")
				os.Exit(-1)
			}
		} else {
			if len(args) > 0 {
				log.Warning("When flag -f given, no arguments will be handled")
			}
		}

		dbFilePath, _, _ := getDbFilePath(cmd)

		switch queryType {
		case "":
			log.Error("Flag -t/--type needed")

		case "gi_taxid_nucl":
			log.Info("Query database: %s", "gi_taxid_nucl")

			if dataFile == "" {
				queryGi2Taxid(dbFilePath, "gi_taxid_nucl", args)
			} else {
				queryGi2TaxidByFile(dbFilePath, "gi_taxid_nucl", dataFile, chunkSize, threads)
			}

		case "gi_taxid_prot":
			log.Info("Query database: %s", "gi_taxid_prot")

			if dataFile == "" {
				queryGi2Taxid(dbFilePath, "gi_taxid_prot", args)
			} else {
				queryGi2TaxidByFile(dbFilePath, "gi_taxid_prot", dataFile, chunkSize, threads)
			}
		default:
			log.Errorf("Unsupported data type: %s", queryType)
			os.Exit(-1)
		}
	},
}

func queryGi2Taxid(dbFilePath string, queryType string, gis []string) {
	db, err := bolt.Open(dbFilePath, 0600, nil)
	defer db.Close()
	checkError(err)

	taxids, err := taxon.QueryGi2Taxid(db, queryType, gis)
	checkError(err)

	for i, gi := range gis {
		fmt.Printf("%s\t%s\n", gi, taxids[i])
	}
}

func queryGi2TaxidByFile(dbFilePath string, queryType string, dataFile string, chunkSize int, threads int) {
	if chunkSize <= 0 {
		chunkSize = 10000
	}
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimSpace(strings.TrimRight(line, "\n"))
		if line == "" {
			return "", false, nil
		}
		return line, true, nil
	}
	reader, err := breader.NewBufferedReader(dataFile, runtime.NumCPU(), chunkSize, fn)
	checkError(err)

	pool := taxon.NewDBPool(dbFilePath, threads)
	chResults := make(chan [][]string, threads)

	// receive result and print
	chDone := make(chan int)
	go func() {
		for s := range chResults {
			gis, taxids := s[0], s[1]
			for i, gi := range gis {
				fmt.Printf("%s\t%s\n", gi, taxids[i])
			}
		}
		chDone <- 1
	}()

	// querying
	var wg sync.WaitGroup
	tokens := make(chan int, threads)
	for chunk := range reader.Ch {
		if chunk.Err != nil {
			checkError(chunk.Err)
			break
		}
		tokens <- 1
		wg.Add(1)

		gis := make([]string, len(chunk.Data))
		for i, data := range chunk.Data {
			gis[i] = data.(string)
		}

		go func(gis []string) {
			db := pool.GetDB()
			defer func() {
				pool.ReleaseDB(db)
				wg.Done()
				<-tokens
			}()

			taxids, err := taxon.QueryGi2Taxid(db, queryType, gis)
			checkError(err)
			chResults <- [][]string{gis, taxids}
		}(gis)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

func init() {
	cliCmd.AddCommand(localCmd)
	localCmd.Flags().StringP("type", "t", "", "query type (see introduction)")
	localCmd.Flags().StringP("file", "f", "", "read queries from file")
	localCmd.Flags().IntP("chunk-size", "c", 100000, "chunk size of querying")
}
