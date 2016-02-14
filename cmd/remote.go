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

	"github.com/shenwei356/gtaxon/taxon"
	fileutil "github.com/shenwei356/util/file"
	"github.com/spf13/cobra"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "query from remote server",
	Long:  ``,
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
		batchSize, err := cmd.Flags().GetInt("batch-size")
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
				remoteQueryGi2TaxidByFile(host, port, "gi_taxid_prot", dataFile, batchSize, threads)
			}

		case "gi_taxid_prot":
			log.Info("Query database: %s from host: %s:%d", "gi_taxid_prot", host, port)

			if dataFile == "" {
				remoteQueryGi2Taxid(host, port, "gi_taxid_prot", args)
			} else {
				remoteQueryGi2TaxidByFile(host, port, "gi_taxid_prot", dataFile, batchSize, threads)
			}

		default:
			log.Errorf("Unsupported data type: %s", dataType)
			os.Exit(-1)
		}
	},
}

func remoteQueryGi2Taxid(host string, port int, dataType string, gis []string) {
	gis, taxids := taxon.RemoteQueryGi2Taxid(host, port, dataType, gis)
	for i, gi := range gis {
		fmt.Printf("%s\t%s\n", gi, taxids[i])
	}
}

func remoteQueryGi2TaxidByFile(host string, port int, dataType string, dataFile string, batchSize int, threads int) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	fn := func(line string) (string, bool) {
		line = strings.TrimSpace(line)
		if line == "" {
			return "", false
		}
		return line, true
	}
	chRead, err := fileutil.ReadFileWithBuffer(dataFile, batchSize, runtime.NumCPU(), fn)
	checkError(err)

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
	for gis := range chRead {
		tokens <- 1
		wg.Add(1)

		go func(gis []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			gis, taxids := taxon.RemoteQueryGi2Taxid(host, port, dataType, gis)
			checkError(err)
			chResults <- [][]string{gis, taxids}
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
	remoteCmd.Flags().StringP("type", "t", "", "data type. Valid: gi_taxid_prot or gi_taxid_nucl")
	remoteCmd.Flags().StringP("file", "f", "", "read queries from file")
	remoteCmd.Flags().IntP("batch-size", "b", 10000, "batch size of querying (should not be too small or too large, do not change this)")
}
