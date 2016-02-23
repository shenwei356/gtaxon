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

// Division defines the NCBI taxonomy division
type Division struct {
	DivisionID   string `json:"DivisionID"`
	DivisionCode string `json:"DivisionCode"`
	DivisionName string `json:"DivisionName"`
	Comments     string `json:"Comments"`
}

// ToJSON get the JSON string of a division
func (division Division) ToJSON() (string, error) {
	s, err := json.Marshal(division)
	return string(s), err
}

// DivisionFromJSON return node object from JSON string
func DivisionFromJSON(s string) (Division, error) {
	var d Division
	err := json.Unmarshal([]byte(s), &d)
	return d, err
}

// DivisionFromArgs is used when importing data from divisions.dmp
func DivisionFromArgs(items []string) Division {
	if len(items) != 4 {
		return Division{}
	}
	return Division{
		DivisionID:   items[0],
		DivisionCode: items[1],
		DivisionName: items[2],
		Comments:     items[3],
	}
}

// Divisions is a map storing all divisions
var Divisions map[string]Division

var mutex3 = &sync.Mutex{}

// SetDivisions sets Nodes
func SetDivisions(divisions map[string]Division) {
	mutex3.Lock()
	Divisions = divisions
	mutex3.Unlock()
}
