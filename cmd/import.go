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
	"os"

	"github.com/shenwei356/gtaxon/taxon"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Importing taxonomy data from files downloaded from NCBI ftp",
	Long: `Importing taxonomy data from files downloaded from NCBI ftp (
ftp://ftp.ncbi.nih.gov/pub/taxonomy).

Supported file types includes:

  ================================================
      data type                  files
  ------------------------------------------------
    gi_taxid_nucl          gi_taxid_nucl.dmp.gz
    gi_taxid_prot          gi_taxid_prot.dmp.gz
    nodes                  nodes.dmp
	names                  names.dmp
    ... to be updated

  ================================================

`,
	Run: func(cmd *cobra.Command, args []string) {
		fileType, err := cmd.Flags().GetString("type")
		checkError(err)

		if len(args) != 1 {
			log.Error("ONE and ONLY ONE file needed")
			os.Exit(-1)
		}
		dataFile := args[0]

		dbFilePath, _, _ := getDbFilePath(cmd)

		chunkSize, err := cmd.Flags().GetInt("chunk-size")
		checkError(err)
		force, err := cmd.Flags().GetBool("force")
		checkError(err)

		switch fileType {
		case "":
			log.Error("Flag -t/--type needed")

		case "gi_taxid_nucl":
			log.Info("Import from file: %s", dataFile)

			taxon.ImportGiTaxid(dbFilePath, "gi_taxid_nucl", dataFile, chunkSize, force)

		case "gi_taxid_prot":
			log.Info("Import from file: %s", dataFile)

			taxon.ImportGiTaxid(dbFilePath, "gi_taxid_prot", dataFile, chunkSize, force)

		case "nodes":
			log.Info("Import from file: %s", dataFile)

			taxon.ImportNodes(dbFilePath, "nodes", dataFile, chunkSize, force)

		case "names":
			log.Info("Import from file: %s", dataFile)

			taxon.ImportNames(dbFilePath, "names", dataFile, chunkSize, force)

		default:
			log.Errorf("Unsupported filetype: %s", fileType)
			os.Exit(-1)
		}
	},
}

func init() {
	dbCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("type", "t", "", "data type. see above description")
	importCmd.Flags().BoolP("force", "f", false, "delete exited subdatabase")
	importCmd.Flags().IntP("chunk-size", "c", 100000, "chunk size of records when writting database (do not change this)")
}
