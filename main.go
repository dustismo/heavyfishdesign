package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/linter"
	"github.com/dustismo/heavyfishdesign/parser"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var FileExtension = "hfd"

func main() {
	// initialize

	// various flags
	renderFilename := flag.String("path", "", "Path to the file to render")
	outputFile := flag.String("output_file", "", "File name to save")

	renderDirectory := flag.String("render_dir", "designs/", "The Directory to render (recursively)")
	outputDirectory := flag.String("output_dir", "", "The Directory to render into")
	compareDirectory := flag.String("compare_dir", "designs_rendered", "The Directory to compare the current render to")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: \n \t$ run main.go [serve|render|render_all|diff_test|designs_updated]\n")
		return
	}
	command := os.Args[1]
	err := flag.CommandLine.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("Error %s", err.Error())
		return
	}

	// the command
	parser.InitContext()

	if command == "serve" {
		http.Handle("/json", http.HandlerFunc(processRequest))
		err := http.ListenAndServe(":2003", nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	} else if command == "render" {
		logger := util.NewLog()
		rfn := os.Args[2]
		if len(*renderFilename) > 0 {
			rfn = *renderFilename
		}

		if len(rfn) == 0 {
			log.Fatalf("Render filename is required")
			return
		}

		planset, err := renderPlanSet(rfn, dynmap.New(), logger)
		if err != nil {
			log.Fatalf("Error during planset render: %s\n", err.Error())
			return
		}

		outFile := createFilename(*outputFile, rfn)
		err = save(planset, outFile, logger)
		if err != nil {
			fmt.Printf("Error during save: %s\n", err.Error())
			return
		}
	} else if command == "render_all" {
		logger := util.NewLog()

		err := RenderAll(*renderDirectory, *outputDirectory, logger)
		if err != nil {
			logger.Errorf("error %s", err.Error())
			return
		}
	} else if command == "designs_updated" {
		logger := util.NewLog()
		// clear out the designs rendered directory
		util.ClearDir("designs_rendered", "svg")
		err := RenderAll("designs", "designs_rendered", logger)
		if err != nil {
			logger.Errorf("error %s", err.Error())
			return
		}
	} else if command == "diff_test" {
		logger := util.NewLog()
		logger.LogToStdOut = util.Fatal

		if outputDirectory == nil || len(*outputDirectory) == 0 {
			od, err := ioutil.TempDir("", "hfd_diff_test")
			if err != nil {
				logger.Errorf("error %s", err.Error())
				return
			}
			outputDirectory = &od
		}
		err := RenderAll(*renderDirectory, *outputDirectory, logger)
		if err != nil {
			fmt.Printf("error %s", err.Error())
			return
		}
		CompareDirectories(*compareDirectory, *outputDirectory, "svg", logger)

		diffCount := 0
		if logger.HasErrors() {
			for _, msg := range logger.Messages {
				if msg.MustBool("bad_diff", false) {
					diffCount = diffCount + 1
					fmt.Printf("*******************\n")
					fmt.Printf("file1: %s\n", msg.MustString("file1", ""))
					fmt.Printf("file2: %s\n", msg.MustString("file2", ""))
					fmt.Printf("diff: \n%s\n", msg.MustString("diff", ""))
					fmt.Printf("*******************\n")
				}
			}
		}

		fmt.Printf("\n\nTotal files with differences: %d\n", diffCount)
		for _, msg := range logger.Messages {
			if msg.MustBool("bad_diff", false) {
				fmt.Printf("\t%s\n", msg.MustString("file1", ""))
			}
		}
	} else if command == "svg" {
		logger := util.NewLog()
		logger.LogToStdOut = util.Info

	} else if command == "lint" {
		logger := util.NewLog()
		logger.LogToStdOut = util.Info
		rfn := os.Args[2]
		if len(*renderFilename) > 0 {
			rfn = *renderFilename
		}
		linter.Lint(rfn, logger)
	} else {
		fmt.Printf("Usage: \n \t$ run main.go [serve|render|render_all|diff_test|designs_updated]\n")
	}
}

func CompareDirectories(left string, right string, extension string, logger *util.HfdLog) {
	filesLeft, err := util.FileList(left, extension)
	if err != nil {
		logger.Errorf("Unable to list files in directory %s: %s", left, err.Error())
		return
	}
	filesRight, err := util.FileList(right, extension)
	if err != nil {
		logger.Errorf("Unable to list files in directory %s: %s", right, err.Error())
		return
	}

	for _, l := range filesLeft {
		loggerL := logger.NewChild()
		loggerL.StaticFields.Put("filename", l)
		_, lBase := filepath.Split(l)

		// now look for the same file base name in the right side
		found := false
		for _, r := range filesRight {
			if !found {
				_, rBase := filepath.Split(r)
				if rBase == lBase {
					// now compare files
					loggerL.Infof("Comparing files:\n\t%s\n\t%s", l, r)
					lBytes, err := ioutil.ReadFile(l)
					if err != nil {
						loggerL.Errorf("error loading %s: %s", l, err.Error())
					}
					rBytes, err := ioutil.ReadFile(r)
					if err != nil {
						loggerL.Errorf("error loading %s: %s", r, err.Error())
					}
					diff := diffmatchpatch.New()
					diffs := diff.DiffMain(string(lBytes), string(rBytes), false)

					if !IsEqualDiff(diffs) {
						loggerL.Errorf("Files not equal: bytes not equal")
						loggerL.Errorfd(dynmap.CreateFromMap(map[string]interface{}{
							"bad_diff": true,
							"diff":     diff.DiffPrettyText(diffs),
							"file1":    l,
							"file2":    r,
						}), diff.DiffPrettyText(diffs))

					} else {
						loggerL.Infof("Files are equal!")
					}
					found = true
				}
			}
		}
		if !found {
			loggerL.Errorf("Unable to find file %s in directory %s", lBase, right)
		}
	}

}
func IsEqualDiff(diffs []diffmatchpatch.Diff) bool {
	eq := true
	for _, d := range diffs {
		if d.Type != diffmatchpatch.DiffEqual {
			eq = false
		}
	}
	return eq
}

// creates the directory listed if it does not already exist
func CreateDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

func RenderAll(renderDir, outputDir string, logger *util.HfdLog) error {
	filenames, err := util.FileList(renderDir, FileExtension)
	if err != nil {
		logger.Errorf("Error during render_all: %s\n", err.Error())
		return err
	}
	err = CreateDir(outputDir)
	if err != nil {
		logger.Errorf("Error creating directory %s: %s\n", outputDir, err.Error())
		return err
	}

	for _, rf := range filenames {
		planLogger := logger.NewChild()
		planLogger.StaticFields.Put("filename", rf)
		planset, err := renderPlanSet(rf, dynmap.New(), planLogger)
		if err != nil {
			planLogger.Errorf("Error during planset %s : %s\n", rf, err.Error())
		} else {
			outFile := createFilename(outputDir, rf)
			err = save(planset, outFile, planLogger)
			if err != nil {
				planLogger.Errorf("Error during save %s : %s\n", rf, err.Error())
			}
		}
	}
	return nil
}

func renderPlanSet(filename string, params *dynmap.DynMap, logger *util.HfdLog) (*dom.PlanSet, error) {
	logger.Infof("RENDERING: %s\n", filename)
	dm, err := linter.LoadDynMap(filename)
	if err != nil {
		return nil, err
	}
	docParams := dm.MustDynMap("params", dynmap.New()).Clone().Merge(params)
	dm.Put("params", docParams)
	doc, err := dom.ParseDocument(dm, logger)
	if err != nil {
		return nil, err
	}
	planset := dom.NewPlanSet(doc)

	context := dom.RenderContext{
		Origin: path.NewPoint(0, 0),
		Cursor: path.NewPoint(0, 0),
	}

	err = planset.Init(context)
	return planset, err
}

// construct a suitable saveFile name from the give path + document name
// for instance:
// filepath = "/home/my_document/"
// docuPath = "/home/designs/box_name.hfd"
// = /home/my_document/box_name_000.svg
func createFilename(filePath string, documentPath string) string {
	_, dcFn := filepath.Split(documentPath)

	// if no filePath is specified then use the documentpath
	if len(filePath) == 0 {
		// take document filename, strip suffix, then use that.
		// local directory in that case.
		return strings.TrimSuffix(dcFn, filepath.Ext(documentPath))
	}

	fpDir, fpFn := filepath.Split(filePath)

	if IsDir(filePath) {
		// need to make sure fbFn is not a directory
		fpDir = filepath.Clean(filePath)
		fpFn = ""
	}

	if len(fpFn) == 0 {
		// filepath directory but not file
		fp := strings.TrimSuffix(dcFn, filepath.Ext(dcFn))
		return filepath.Join(fpDir, fp)
	}

	return strings.TrimSuffix(filePath, filepath.Ext(filePath))
}

// see if this path is an existing directory
func IsDir(filePath string) bool {
	f, err := os.Open(filePath)
	if err == nil {
		defer f.Close()
		stat, err := f.Stat()
		if err == nil {
			return stat.IsDir()
		}
	}
	return false
}

func save(planset *dom.PlanSet, saveFile string, logger *util.HfdLog) error {
	svgDocs := planset.SVGDocuments()
	context := dom.RenderContext{
		Origin: path.NewPoint(0, 0),
		Cursor: path.NewPoint(0, 0),
		Log:    logger,
	}

	for i, svgDoc := range svgDocs {
		fn := fmt.Sprintf("%s_%03d.svg", saveFile, i)
		f, err := os.Create(fn)
		if err != nil {
			return err
		}
		defer f.Close()
		logger.Infof("SAVING: %s\n", fn)
		svgDoc.WriteSVG(context, f)
	}
	return nil
}

func processRequest(w http.ResponseWriter, req *http.Request) {
	params := dynmap.New()
	err := req.ParseForm()
	if err != nil {
		println(err.Error())
		return
	}
	params.UnmarshalUrlValues(req.Form)
	logger := util.NewLog()
	filename := params.MustString("file", "dom/testdata/box_test.hfd")
	planset, err := renderPlanSet(filename, params, logger)
	if err != nil {
		logger.Errorf("Error during render: %s\n", err.Error())
		return
	}
	svgDocs := planset.SVGDocuments()
	context := dom.RenderContext{
		Origin: path.NewPoint(0, 0),
		Cursor: path.NewPoint(0, 0),
	}
	svgDocs[params.MustInt("index", 0)].WriteSVG(context, w)
	w.Header().Set("Content-Type", "image/svg+xml")

	saveFile, ok := params.GetString("save_file")
	if ok {
		sf := createFilename(filename, saveFile)
		err = save(planset, sf, logger)
		if err != nil {
			logger.Errorf("Error during save: %s\n", err.Error())
			return
		}
	}
}
