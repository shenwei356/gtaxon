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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/shenwei356/breader"
	"github.com/shenwei356/gtaxon/taxon"
	"github.com/spf13/cobra"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "query from remote server",
	Long: `query from remote server.

Available query types:

    gi_taxid_nucl      query TaxId by Gi (nucl)
    gi_taxid_prot      query TaxId by Gi (prot)

    taxid2taxon        query Taxon by TaxId
    name2taxid         query TaxId by Name
    lca                query Lowest Common Ancestor by TaxIds

`,
	Run: func(cmd *cobra.Command, args []string) {
		runtime.GOMAXPROCS(runtime.NumCPU())

		dataType, err := cmd.Flags().GetString("type")
		checkError(err)
		host, err := cmd.Flags().GetString("host")
		checkError(err)
		port, err := cmd.Flags().GetInt("port")
		checkError(err)
		dataFile, err := cmd.Flags().GetString("file")
		checkError(err)
		chunkSize, err := cmd.Flags().GetInt("chunk-size")
		checkError(err)
		threads, err := cmd.Flags().GetInt("threads")
		checkError(err)

		if dataFile == "" {
			if len(args) == 0 {
				log.Error("Queries needed. Type \"gtaxon cli remote -h\" for help")
				os.Exit(-1)
			}
		} else {
			if len(args) > 0 {
				log.Warning("When flag -f given, no arguments will be handled")
			}
		}

		switch dataType {
		case "":
			log.Error("Flag -t/--type needed")

		case "gi_taxid_nucl":
			log.Info("Query database: %s from host: %s:%s", "gi_taxid_nucl", host, port)

			if dataFile == "" {
				remoteQueryGi2Taxid(host, port, "gi_taxid_prot", args)
			} else {
				remoteQueryGi2TaxidByFile(host, port, "gi_taxid_prot", dataFile, chunkSize, threads)
			}

		case "gi_taxid_prot":
			log.Info("Query database: %s from host: %s:%d", "gi_taxid_prot", host, port)

			if dataFile == "" {
				remoteQueryGi2Taxid(host, port, "gi_taxid_prot", args)
			} else {
				remoteQueryGi2TaxidByFile(host, port, "gi_taxid_prot", dataFile, chunkSize, threads)
			}

		case "taxid2taxon":
			log.Info("Query Taxon by TaxId from host: %s:%d", host, port)

			if dataFile == "" {
				remoteQueryTaxid2Taxon(host, port, args)
			} else {
				remoteQueryTaxid2TaxonByFile(host, port, dataFile, chunkSize, threads)
			}

		case "name2taxid":
			log.Info("Query TaxId by Name from host: %s:%d", host, port)

			nameClass, err := cmd.Flags().GetString("name-class")
			checkError(err)
			useRegexp, err := cmd.Flags().GetBool("use-regexp")
			checkError(err)
			if dataFile == "" {
				remoteName2TaxID(host, port, useRegexp, nameClass, args)
			} else {
				remoteName2TaxIDByFile(host, port, useRegexp, nameClass, dataFile, chunkSize, threads)
			}

		case "lca":
			log.Info("Query LCA by TaxIds from host: %s:%d", host, port)

			if dataFile == "" {
				remoteQueryLCA(host, port, args)
			} else {
				remoteQueryLCAByFile(host, port, dataFile, chunkSize, threads)
			}

		default:
			log.Errorf("Unsupported data type: %s", dataType)
			os.Exit(-1)
		}
	},
}

// --------------------------------------------------------------------------

func remoteQueryLCA(host string, port int, queries []string) {
	msg := taxon.RemoteQueryLCA(host, port, queries)
	if msg.Status != "OK" {
		log.Error(msg.Message)
	}

	for query, taxon := range msg.LCA {
		fmt.Printf("Query TaxIDs: %s\n", query)
		bs, err := json.MarshalIndent(taxon, "", "  ")
		checkError(err)
		fmt.Printf("Taxon: %s\n\n", string(bs))
	}
}

func remoteQueryLCAByFile(host string, port int, dataFile string, chunkSize int, threads int) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimRight(line, "\n")
		if line == "" {
			return "", false, nil
		}
		return line, true, nil
	}
	reader, err := breader.NewBufferedReader(dataFile, threads, chunkSize, fn)
	checkError(err)

	chResults := make(chan taxon.MessageLCAMap, threads)

	// receive result and print
	chDone := make(chan int)
	go func() {
		for msg := range chResults {
			if msg.Status != "OK" {
				log.Error(msg.Message)
			}
			for query, taxon := range msg.LCA {
				fmt.Printf("Query TaxIDs: %s\n", query)
				bs, err := json.MarshalIndent(taxon, "", "  ")
				checkError(err)
				fmt.Printf("Taxon: %s\n\n", string(bs))
			}
		}
		chDone <- 1
	}()

	// querying
	var wg sync.WaitGroup
	tokens := make(chan int, threads)
	for chunk := range reader.Ch {
		tokens <- 1
		wg.Add(1)

		queries := make([]string, len(chunk.Data))
		for i, data := range chunk.Data {
			queries[i] = data.(string)
		}

		go func(queries []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			msg := taxon.RemoteQueryLCA(host, port, queries)
			checkError(err)
			chResults <- msg
		}(queries)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

// --------------------------------------------------------------------------

func remoteQueryTaxid2Taxon(host string, port int, taxids []string) {
	msg := taxon.RemoteQueryTaxid2Taxon(host, port, taxids)
	if msg.Status != "OK" {
		log.Error(msg.Message)
	}

	for taxid, taxon := range msg.Taxons {
		fmt.Printf("Query TaxIDs: %s\n", taxid)
		bs, err := json.MarshalIndent(taxon, "", "  ")
		checkError(err)
		fmt.Printf("Taxon: %s\n\n", string(bs))
	}
}

func remoteQueryTaxid2TaxonByFile(host string, port int, dataFile string, chunkSize int, threads int) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimRight(line, "\n")
		if line == "" {
			return "", false, nil
		}
		return line, true, nil
	}
	reader, err := breader.NewBufferedReader(dataFile, threads, chunkSize, fn)
	checkError(err)

	chResults := make(chan taxon.MessageTaxid2TaxonMap, threads)

	// receive result and print
	chDone := make(chan int)
	go func() {
		for msg := range chResults {
			if msg.Status != "OK" {
				log.Error(msg.Message)
			}
			for taxid, taxon := range msg.Taxons {
				fmt.Printf("Query TaxIDs: %s\n", taxid)
				bs, err := json.MarshalIndent(taxon, "", "  ")
				checkError(err)
				fmt.Printf("Taxon: %s\n\n", string(bs))
			}
		}
		chDone <- 1
	}()

	// querying
	var wg sync.WaitGroup
	tokens := make(chan int, threads)
	for chunk := range reader.Ch {
		tokens <- 1
		wg.Add(1)

		queries := make([]string, len(chunk.Data))
		for i, data := range chunk.Data {
			queries[i] = data.(string)
		}

		go func(queries []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			msg := taxon.RemoteQueryTaxid2Taxon(host, port, queries)
			checkError(err)
			chResults <- msg
		}(queries)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

// --------------------------------------------------------------------------

func remoteName2TaxID(host string, port int, useRegexp bool, nameClass string, names []string) {
	msg := taxon.RemoteQueryName2TaxID(host, port, useRegexp, nameClass, names)
	if msg.Status != "OK" {
		log.Error(msg.Message)
	}

	for name, items := range msg.TaxIDs {
		idnames := make([]string, len(items))
		for i, item := range items {
			idnames[i] = fmt.Sprintf("%d(%s)", item.TaxID, item.ScientificName)
		}
		fmt.Printf("%s\t%s\n", name, strings.Join(idnames, ","))
	}
}

func remoteName2TaxIDByFile(host string, port int, useRegexp bool, nameClass string, dataFile string, chunkSize int, threads int) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimRight(line, "\n")
		if line == "" {
			return "", false, nil
		}
		return line, true, nil
	}
	reader, err := breader.NewBufferedReader(dataFile, threads, chunkSize, fn)
	checkError(err)

	chResults := make(chan taxon.MssageName2TaxIDMap, threads)

	// receive result and print
	chDone := make(chan int)
	go func() {
		for msg := range chResults {
			if msg.Status != "OK" {
				log.Error(msg.Message)
			}
			for name, items := range msg.TaxIDs {
				idnames := make([]string, len(items))
				for i, item := range items {
					idnames[i] = fmt.Sprintf("%d(%s)", item.TaxID, item.ScientificName)
				}
				fmt.Printf("%s\t%s\n", name, strings.Join(idnames, ","))
			}
		}
		chDone <- 1
	}()

	// querying
	var wg sync.WaitGroup
	tokens := make(chan int, threads)
	for chunk := range reader.Ch {
		tokens <- 1
		wg.Add(1)

		queries := make([]string, len(chunk.Data))
		for i, data := range chunk.Data {
			queries[i] = data.(string)
		}

		go func(queries []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			msg := taxon.RemoteQueryName2TaxID(host, port, useRegexp, nameClass, queries)
			checkError(err)
			chResults <- msg
		}(queries)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

// --------------------------------------------------------------------------

func remoteQueryGi2Taxid(host string, port int, dataType string, gis []string) {
	msg := taxon.RemoteQueryGi2Taxid(host, port, dataType, gis)
	if msg.Status != "OK" {
		log.Error(msg.Message)
	}
	for gi, taxid := range msg.Taxids {
		fmt.Printf("%s\t%s\n", gi, taxid)
	}
}

func remoteQueryGi2TaxidByFile(host string, port int, dataType string, dataFile string, chunkSize int, threads int) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimSpace(strings.TrimRight(line, "\n"))
		if line == "" {
			return "", false, nil
		}
		return line, true, nil
	}
	reader, err := breader.NewBufferedReader(dataFile, threads, chunkSize, fn)
	checkError(err)

	chResults := make(chan taxon.MessageGI2TaxidMap, threads)

	// receive result and print
	chDone := make(chan int)
	go func() {
		for msg := range chResults {
			if msg.Status != "OK" {
				log.Error(msg.Message)
			}
			for gi, taxid := range msg.Taxids {
				fmt.Printf("%s\t%s\n", gi, taxid)
			}
		}
		chDone <- 1
	}()

	// querying
	var wg sync.WaitGroup
	tokens := make(chan int, threads)
	for chunk := range reader.Ch {
		tokens <- 1
		wg.Add(1)

		gis := make([]string, len(chunk.Data))
		for i, data := range chunk.Data {
			gis[i] = data.(string)
		}

		go func(gis []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			msg := taxon.RemoteQueryGi2Taxid(host, port, dataType, gis)
			checkError(err)
			chResults <- msg
		}(gis)
	}
	wg.Wait()
	close(chResults)
	<-chDone
}

func init() {
	cliCmd.AddCommand(remoteCmd)

	remoteCmd.Flags().StringP("host", "H", "127.0.0.1", "server host")
	remoteCmd.Flags().IntP("port", "P", 8080, "port number")
	remoteCmd.Flags().StringP("type", "t", "", `query type. type "gataxon cli remote -h" for help`)
	remoteCmd.Flags().StringP("file", "f", "", "read queries from file")
	remoteCmd.Flags().IntP("chunk-size", "c", 10000, "chunk size of querying (should not be too small or too large, do not change this)")
	remoteCmd.Flags().BoolP("use-regexp", "R", false, `use regexp (only for query type "name2taxid")`)
	remoteCmd.Flags().StringP("name-class", "C", "", `name class (only for query type "name2taxid")`)
}
