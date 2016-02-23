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

//Package nodes a
package nodes

import (
	"encoding/json"
	"sync"
)

// GenCode defines the NCBI taxonomy division
type GenCode struct {
	GenCodeID        string `json:"GenCodeID"`
	Abbreviation     string `json:"Abbreviation"`
	Name             string `json:"Name"`
	TranslationTable string `json:"TranslationTable"`
	StartCodons      string `json:"StartCodons"`
}

// ToJSON get the JSON string of a gencode
func (gencode GenCode) ToJSON() (string, error) {
	s, err := json.Marshal(gencode)
	return string(s), err
}

// GenCodeFromJSON return node object from JSON string
func GenCodeFromJSON(s string) (GenCode, error) {
	var gencode GenCode
	err := json.Unmarshal([]byte(s), &gencode)
	return gencode, err
}

// GenCodeFromArgs is used when importing data from divisions.dmp
func GenCodeFromArgs(items []string) GenCode {
	if len(items) != 5 {
		return GenCode{}
	}
	return GenCode{
		GenCodeID:        items[0],
		Abbreviation:     items[1],
		Name:             items[2],
		TranslationTable: items[3],
		StartCodons:      items[4],
	}
}

// GenCodes is a map storing all divisions
var GenCodes map[string]GenCode

var mutex4 = &sync.Mutex{}

// SetGenCodes sets GenCodes
func SetGenCodes(divisions map[string]GenCode) {
	mutex4.Lock()
	GenCodes = divisions
	mutex4.Unlock()
}
