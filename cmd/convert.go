/*
Copyright © 2022 5km <5km@smslit.cn>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/smslit/datasetgo/model"
	"github.com/spf13/cobra"
)

// Define the dataset format enums
type DatasetFormat string

const (
	COCO      DatasetFormat = "coco"
	PascalVOC DatasetFormat = "voc"
	CreateML  DatasetFormat = "createml"
)

// the format of the source dataset
var iFormat DatasetFormat

// the format of the outputed dataset
var oFormat DatasetFormat

// the path of the outputed dataset, a file or directory
var oDatasetPath string

// the path of the source dataset, a file or directory
var datasetPath string

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [flags] dataset-path",
	Short: "A subcommand to convert the dataset format",
	Long: `A subcommand to convert the dataset format. The supported 
formats as follows:
- coco: COCO
- voc: PascalVOC
- createml: Create ML(apple)`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		// check the path if exist
		if _, err := os.Stat(args[0]); err == nil || os.IsExist(err) {
			datasetPath = args[0]
			return nil
		}

		return errors.New("the dataset-path does not exist")
	},
	Run: func(cmd *cobra.Command, args []string) {

		switch oFormat {
		case PascalVOC:
			ConvertToPascalVOC(iFormat, oFormat, datasetPath, oDatasetPath)

		case COCO:
			ConvertToCOCO(iFormat, oFormat, datasetPath, oDatasetPath)

		case CreateML:
			ConvertToCreateML(iFormat, oFormat, datasetPath, oDatasetPath)

		default:
			rootCmd.PrintErrln(errors.New("the specified format is not supported"))
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP((*string)(&iFormat), "input-format", "i", "", "the format of the source dataset")
	convertCmd.MarkFlagRequired("iutput-format")
	convertCmd.Flags().StringVarP((*string)(&oFormat), "output-format", "o", "", "the format of the outputed dataset")
	convertCmd.MarkFlagRequired("output-format")
	convertCmd.Flags().StringVarP((*string)(&oDatasetPath), "output-path", "p", "", "the path of the outputed dataset, a file or directory")
}

func ConvertToPascalVOC(iFormat DatasetFormat, oFormat DatasetFormat, datasetPath string, oDatasetPath string) {
	var err error
	var annotations model.VOCAnnotations

	if iFormat == COCO {
		err = model.ReadVOCAnnotationsFromCOCOFile(&annotations, datasetPath)
	} else {
		err = model.ReadVOCAnnotationsFromCreateMLFile(&annotations, datasetPath)
	}

	if err != nil {
		rootCmd.PrintErrln(err)
		return
	}

	// get an valid output path
	if oDatasetPath == "" {
		oDatasetPath = filepath.Dir(datasetPath)
	} else {
		if _, err := os.Stat(oDatasetPath); !(err == nil || os.IsExist(err)) {
			os.MkdirAll(oDatasetPath, os.ModePerm)
		}
	}

	// output the annotations data to xml files
	if err := model.WriteVOCAnnotationsToFile(&annotations, oDatasetPath); err != nil {
		rootCmd.PrintErrln(err)
	}
}

func ConvertToCOCO(iFormat DatasetFormat, oFormat DatasetFormat, datasetPath string, oDatasetPath string) {
	var err error
	var dataDir string
	var annotations model.COCOAnnotations

	if iFormat == PascalVOC {
		dataDir = datasetPath
		err = model.ReadCOCOAnnotationsFromPascalVOCDir(&annotations, datasetPath)
	} else {
		dataDir = filepath.Dir(datasetPath)
		err = model.ReadCOCOAnnotationsFromCreateMLFile(&annotations, datasetPath)
	}

	if err != nil {
		rootCmd.PrintErrln(err)
		return
	}

	// get an valid output path
	if oDatasetPath == "" {
		nowTimeString := time.Now().Format("20060102150405")
		oDatasetPath = filepath.Join(dataDir, fmt.Sprintf("_annotations.coco.%v.json", nowTimeString))
	} else {
		if pathExt := filepath.Ext(oDatasetPath); pathExt == "" || strings.ToLower(pathExt) != ".json" {
			rootCmd.PrintErrln(errors.New(oDatasetPath + " is not a valid json file path"))
			return
		}
	}

	// output the valid coco json file
	if err := model.WriteCOCOAnnotationsToFile(&annotations, oDatasetPath); err != nil {
		rootCmd.PrintErrln(err)
	}
}

func ConvertToCreateML(iFormat DatasetFormat, oFormat DatasetFormat, datasetPath string, oDatasetPath string) {
	var err error
	var dataDir string
	var annotations model.CreateMLAnnotations
	if iFormat == PascalVOC {
		dataDir = datasetPath
		err = model.ReadCreateMLAnnotationsFromPascalVOCDir(&annotations, datasetPath)
	} else {
		dataDir = filepath.Dir(datasetPath)
		err = model.ReadCreateMLAnnotationsFromCOCOFile(&annotations, datasetPath)
	}

	if err != nil {
		rootCmd.PrintErrln(err)
		return
	}

	// get an valid output path
	if oDatasetPath == "" {
		nowTimeString := time.Now().Format("20060102150405")
		oDatasetPath = filepath.Join(dataDir, fmt.Sprintf("_annotations.createml.%v.json", nowTimeString))
	} else {
		if pathExt := filepath.Ext(oDatasetPath); pathExt == "" || strings.ToLower(pathExt) != ".json" {
			rootCmd.PrintErrln(errors.New(oDatasetPath + " is not a valid json file path"))
			return
		}
	}

	if err := model.WriteCreateMLAnnotationsToFile(&annotations, oDatasetPath); err != nil {
		rootCmd.PrintErrln(err)
	}
}
