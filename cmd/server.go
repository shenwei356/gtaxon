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
	"runtime"

	"github.com/spf13/cobra"

	"github.com/shenwei356/gtaxon/taxon"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start a web server",
	Long: `start a web server with REST APIs,
and you can use "gtaxon cli remote" to query from the server `,
	Run: func(cmd *cobra.Command, args []string) {
		dbFilePath, _, _ := getDbFilePath(cmd)

		port, err := cmd.Flags().GetInt("port")
		checkError(err)
		timeout, err := cmd.Flags().GetInt("timeout")
		checkError(err)
		threads, err := cmd.Flags().GetInt("threads")
		checkError(err)

		taxon.StartServer(dbFilePath, port, timeout, threads)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntP("port", "P", 8080, "port number")
	serverCmd.Flags().IntP("timeout", "", 2, "request and responce time out (second)")
	serverCmd.Flags().IntP("threads", "j", runtime.NumCPU(), "max number of database connection")
}
