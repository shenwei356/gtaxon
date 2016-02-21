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

import "encoding/json"

// NameItem defines
type NameItem struct {
	Name       string `json:"Name"`
	UniqueName string `json:"UniqueName"`
	NameClass  string `json:"NameClass"`
}

// Name includes all names for a taxid
type Name struct {
	TaxID string     `json:"TaxID"`
	Names []NameItem `json:"Names"`
}

// ToJSON transforms Name to JSON string
func (name Name) ToJSON() (string, error) {
	s, err := json.Marshal(name)
	return string(s), err
}

// NameFromJSON return Name object from JSON string
func NameFromJSON(s string) (Name, error) {
	var name Name
	err := json.Unmarshal([]byte(s), &name)
	return name, err
}

// NameFromArgs is used when importing data from names.dmp
func NameFromArgs(items []string) Name {
	if len(items) != 4 {
		return Name{}
	}

	return Name{
		TaxID: items[0],
		Names: []NameItem{
			NameItem{
				Name:       items[1],
				UniqueName: items[2],
				NameClass:  items[3],
			}},
	}
}

// MergeNames is used when importing data from names.dmp
func MergeNames(names ...Name) Name {
	if len(names) < 2 {
		return names[0]
	}
	name := names[0]
	for _, another := range names[1:] {
		if another.TaxID != name.TaxID {
			continue
		}
		for _, anotherNameItem := range another.Names {
			name.Names = append(name.Names, anotherNameItem)
		}
	}
	return name
}

// Names is a map storing all names
var Names map[string]Name
