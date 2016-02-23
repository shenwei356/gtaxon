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
	"fmt"
	"strconv"
	"strings"
)

// GetTaxonByTaxID return Taxon obejct by taxid
func GetTaxonByTaxID(taxid string) (Taxon, error) {
	taxon := Taxon{}

	var node Node
	if _, ok := Nodes[taxid]; !ok {
		return taxon, fmt.Errorf("no Node matches taxid: %s", taxid)
	}
	node = Nodes[taxid]

	var name Name
	if _, ok := Names[taxid]; !ok {
		return taxon, fmt.Errorf("no Name matches taxid: %s", taxid)
	}
	name = Names[taxid]

	var division Division
	if _, ok := Divisions[node.DivisionID]; !ok {
		return taxon, fmt.Errorf("no Division matches division id: %s", node.DivisionID)
	}
	division = Divisions[node.DivisionID]

	var gencode, mgencode GenCode
	if _, ok := GenCodes[node.GeneticCodeID]; !ok {
		return taxon, fmt.Errorf("no GenCode matches genetic code id: %s", node.GeneticCodeID)
	}
	gencode = GenCodes[node.GeneticCodeID]
	mgencode = GenCodes[node.MitochondrialGCID]

	taxon.TaxId, _ = strconv.Atoi(node.TaxID)
	taxon.ParentTaxId, _ = strconv.Atoi(node.PTaxID)
	taxon.Rank = node.Rank
	taxon.Division = division.DivisionName

	taxon.OtherNames = []TaxonNameItem{}
	for _, nameItem := range name.Names {
		if nameItem.NameClass == "scientific name" {
			taxon.ScientificName = nameItem.Name
		} else {
			taxon.OtherNames = append(taxon.OtherNames, TaxonNameItem{
				ClassCDE: nameItem.NameClass,
				DispName: nameItem.Name,
			})
		}
	}

	gcid, _ := strconv.Atoi(gencode.GenCodeID)
	taxon.GeneticCode = GeneticCodeItem{
		GCId:   gcid,
		GCName: gencode.Name,
	}
	mgcid, _ := strconv.Atoi(mgencode.GenCodeID)
	taxon.MitoGeneticCode = MitoGeneticCodeItem{
		MGCId:   mgcid,
		MGCName: mgencode.Name,
	}

	ancestors := ancestorsOfNode(Nodes, node)
	if len(ancestors) <= 2 {
		return taxon, nil
	}
	lineageExItems := make([]LineageExItem, len(ancestors)-2)
	LineageNameSlice := make([]string, len(ancestors)-2)
	j := 0
	for i := len(ancestors) - 2; i >= 1; i-- { // exclude root node and itself
		anc := ancestors[i]
		taxidInt, _ := strconv.Atoi(anc.TaxID)
		scientificName := ""
		for _, nameItem := range Names[anc.TaxID].Names {
			if nameItem.NameClass == "scientific name" {
				scientificName = nameItem.Name
			}
		}
		lineageExItems[j] = LineageExItem{
			TaxId:          taxidInt,
			ScientificName: scientificName,
			Rank:           anc.Rank,
		}
		LineageNameSlice[j] = scientificName
		j++
	}
	taxon.Lineage = strings.Join(LineageNameSlice, "; ")
	taxon.LineageEx = lineageExItems
	return taxon, nil
}

// Taxon is for json output
type Taxon struct {
	TaxId           int `json:"TaxId"`
	ScientificName  string
	OtherNames      []TaxonNameItem
	ParentTaxId     int
	Rank            string
	Division        string
	GeneticCode     GeneticCodeItem
	MitoGeneticCode MitoGeneticCodeItem
	Lineage         string
	LineageEx       []LineageExItem
}

// TaxonNameItem is
type TaxonNameItem struct {
	ClassCDE string
	DispName string
}

// GeneticCodeItem is
type GeneticCodeItem struct {
	GCId   int
	GCName string
}

// MitoGeneticCodeItem is
type MitoGeneticCodeItem struct {
	MGCId   int
	MGCName string
}

// LineageExItem is
type LineageExItem struct {
	TaxId          int
	ScientificName string
	Rank           string
}

func (taxon Taxon) ToJSON() (string, error) {
	s, err := json.Marshal(taxon)
	return string(s), err
}
