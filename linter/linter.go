package linter

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/util"
)

func LoadDynMap(filename string) (*dynmap.DynMap, error) {
	b, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		return nil, err
	}
	json := string(b) // convert content to a 'string'

	dm, err := dynmap.ParseJSON(json)
	return dm, err
}

func lintComponent(comp *dynmap.DynMap) *dynmap.DynMap {
	return comp
}

func extractParams(mp *dynmap.DynMap) *dynmap.DynMap {
	denyList := []string{
		"custom_component",
		"id",
		"transforms",
		"components",
		"component",
		"type",
		"commands",
		"command",
		"repeat",
		"defaults", // deprecated
	}

	tmp := mp.Clone()
	tmp.RemoveAll(denyList...)
	pms := tmp.MustDynMap("params", dynmap.New())
	tmp.Remove("params")
	tmp = tmp.Merge(pms)

	params, _ := dom.ParseParams(tmp)
	newParams := dynmap.New()
	for _, p := range params {
		hfdP := p.ToHFD()
		if hfdP.Length() == 1 {
			// only contains value?
			newParams.Put(p.Key, p.Value)
		} else {
			newParams.Put(p.Key, p.ToHFD())
		}
	}
	return newParams
}

func lintPart(part *dynmap.DynMap) *dynmap.DynMap {

	newPart := dynmap.New()

	newPart.Put("id", part.MustString("id", newId()))
	newPart.Put("type", part.MustString("type", ""))

	customComponent := part.MustDynMap("custom_component", dynmap.New())
	if customComponent.Length() > 0 {
		newPart.Put("custom_component", customComponent)
	}

	transforms := part.MustDynMapSlice("transforms", []*dynmap.DynMap{})
	if len(transforms) > 0 {
		newPart.Put("transforms", transforms)
	}

	//  move attributes to params
	params := extractParams(part)
	if params.Length() > 0 {
		newPart.Put("params", params)
	}

	components := []*dynmap.DynMap{}
	for _, c := range part.MustDynMapSlice("components", []*dynmap.DynMap{}) {
		comp := lintPart(c)
		components = append(components, comp)
	}
	if len(components) > 0 {
		newPart.Put("components", components)
	}
	return newPart
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func newId() string {
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Lint(filename string, logger *util.HfdLog) error {
	if strings.HasSuffix(filename, ".hfd") {
		return LintFile(filename, true, logger)
	} else {
		// is a directory?
		filenames, err := util.FileList(filename, "hfd")
		if err != nil {
			return err
		}
		for _, fn := range filenames {
			fmt.Printf("LINTING: %s\n", fn)
			err = LintFile(fn, true, logger)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func LintFile(filename string, save bool, logger *util.HfdLog) error {
	// first render it to make sure it is valid
	dm, err := LoadDynMap(filename)

	// normalize params
	params, _ := dom.ParseParams(dm.MustDynMap("params", dynmap.New()))
	newParams := dynmap.New()
	for _, p := range params {
		newParams.Put(p.Key, p.ToHFD())
	}
	dm.Put("params", newParams)

	parts := []*dynmap.DynMap{}
	for _, p := range dm.MustDynMapSlice("parts", []*dynmap.DynMap{}) {
		parts = append(parts, lintPart(p))
	}

	dm.Put("parts", parts)
	println(dm.ToJSON())
	if true {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(dm.ToJSON())
	}
	return err
}
