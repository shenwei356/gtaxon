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
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd is RootCmd
var RootCmd = &cobra.Command{
	Use:   "gtaxon",
	Short: "gtaxon xxxx",
	Long: `gtaxon - a fast cross-platform NCBI taxonomy data querying tool,
with cmd client ans REST API server for both local and remote server.

Version: V0.2
Detail: http://github.com/shenwei356/gtaxon
`,
}

// Execute executes Execute
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

var homeDir string

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("db-dir", "", defaultDbPath, "database directory")
	RootCmd.PersistentFlags().StringP("db-file", "", defaultDbFile, "database file name")

	var err error
	homeDir, err = homedir.Expand("~")
	checkError(err)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", filepath.Join(homeDir, ".gtaxon.yaml"), "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".gtaxon") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	// viper.AutomaticEnv()            // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
