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
	"errors"
	"sync"

	"github.com/shenwei356/util/stringutil"
)

// Node defines the NCBI taxonomy node
type Node struct {
	TaxID  string `json:"TaxID"`
	PTaxID string `json:"PTaxID"`
	Rank   string `json:"Rank"`

	EMBLCode string `json:"EMBLCode"`

	DivisionID       string `json:"DivisionID"`
	InheritedDivFlag bool   `json:"InheritedDivFlag"`

	GeneticCodeID   string `json:"GeneticCodeID"`
	InheritedGCFlag bool   `json:"InheritedGCFlag"`

	MitochondrialGCID string `json:"MitochondrialGCID"`
	InheritedMGCFlag  bool   `json:"InheritedMGCFlag"`

	GenBankHiddenFlag     bool `json:"GenBankHiddenFlag"`
	HiddenSubtreeRootFlag bool `json:"HiddenSubtreeRootFlag"`

	Comments string `json:"Comments"`

	Names []NameItem `json:"Names"`
}

// ToJSON get the JSON string of a node
func (node Node) ToJSON() (string, error) {
	s, err := json.Marshal(node)
	return string(s), err
}

// NodeFromJSON return node object from JSON string
func NodeFromJSON(s string) (Node, error) {
	var node Node
	err := json.Unmarshal([]byte(s), &node)
	return node, err
}

// NodeFromArgs is used when importing data from nodes.dmp
func NodeFromArgs(items []string) Node {
	if len(items) != 13 {
		return Node{}
	}
	TaxID := items[0]
	PTaxID := items[1]
	Rank := items[2]
	EMBLCode := items[3]
	DivisionID := items[4]
	var InheritedDivFlag bool
	if items[5] == "1" {
		InheritedDivFlag = true
	}
	GeneticCodeID := items[6]
	var InheritedGCFlag bool
	if items[7] == "1" {
		InheritedGCFlag = true
	}
	MitochondrialGCID := items[8]
	var InheritedMGCFlag bool
	if items[9] == "1" {
		InheritedMGCFlag = true
	}
	var GenBankHiddenFlag bool
	if items[10] == "1" {
		GenBankHiddenFlag = true
	}
	var HiddenSubtreeRootFlag bool
	if items[11] == "1" {
		HiddenSubtreeRootFlag = true
	}
	Comments := items[12]
	return Node{
		TaxID:                 TaxID,
		PTaxID:                PTaxID,
		Rank:                  Rank,
		EMBLCode:              EMBLCode,
		DivisionID:            DivisionID,
		InheritedDivFlag:      InheritedDivFlag,
		GeneticCodeID:         GeneticCodeID,
		InheritedGCFlag:       InheritedGCFlag,
		MitochondrialGCID:     MitochondrialGCID,
		InheritedMGCFlag:      InheritedMGCFlag,
		GenBankHiddenFlag:     GenBankHiddenFlag,
		HiddenSubtreeRootFlag: HiddenSubtreeRootFlag,
		Comments:              Comments,
	}
}

// Nodes is a map storing all nodes
var Nodes map[string]Node

var mutex2 = &sync.Mutex{}

// SetNodes sets Nodes
func SetNodes(nodes map[string]Node) {
	mutex2.Lock()
	Nodes = nodes
	mutex2.Unlock()
}

// LCA return the lowest common ancestor for a list of taxids
func LCA(nodes map[string]Node, taxids []string) (Node, error) {
	if Nodes == nil {
		return Node{}, errors.New("nodes is nil")
	}
	if len(taxids) < 2 {
		return Node{}, errors.New(">=2 taxids needed")
	}

	currents := make(map[string]Node)
	for _, taxid := range taxids {
		if _, ok := nodes[taxid]; !ok {
			continue
		}
		currents[taxid] = nodes[taxid]
	}

	allAncestors := [][]Node{}
	for _, node := range currents {
		allAncestors = append(allAncestors, ancestorsOfNode(nodes, node))
	}

	commonAncestors := make(map[string]int)
	for _, ancestors := range allAncestors {
		if len(commonAncestors) == 0 {
			i := 1
			for _, ancestor := range ancestors {
				commonAncestors[ancestor.TaxID] = i
				i++
			}
		} else {
			existed := []string{}
			for _, ancestor := range ancestors {
				if _, ok := commonAncestors[ancestor.TaxID]; ok {
					existed = append(existed, ancestor.TaxID)
				}
			}
			newCommonAncestors := make(map[string]int)
			for _, taxid := range existed {
				newCommonAncestors[taxid] = commonAncestors[taxid]
			}
			commonAncestors = newCommonAncestors
		}
	}

	sorted := stringutil.SortCountOfString(commonAncestors, false)
	return nodes[sorted[0].Key], nil
}

func ancestorsOfNode(nodes map[string]Node, node Node) []Node {
	current := node
	parrent := nodes[current.PTaxID]
	ancestors := []Node{current}

	for parrent.TaxID != "1" {
		ancestors = append(ancestors, parrent)
		current = parrent
		parrent = nodes[current.PTaxID]
	}

	if current.TaxID != "1" {
		ancestors = append(ancestors, parrent) // root
		return ancestors
	}

	return ancestors
}
