package model

import (
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type VOCDataSource struct {
	Database   string `xml:"database"`
	Annotation string `xml:"annotation"`
	Image      string `xml:"image"`
}

type VOCOwner struct {
	Name string `xml:"name"`
}

type VOCImageSize struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
	Depth  int `xml:"depth"`
}

type VOCPose string

const (
	Front       VOCPose = "front"
	Rear        VOCPose = "rear"
	Left        VOCPose = "left"
	Right       VOCPose = "right"
	Unspecified VOCPose = "unspecified"
)

type VOCBndbox struct {
	Xmin int `xml:"xmin"`
	Xmax int `xml:"xmax"`
	Ymin int `xml:"ymin"`
	Ymax int `xml:"ymax"`
}

type VOCAnnotationItem struct {
	Name      string    `xml:"name"`
	Pose      VOCPose   `xml:"pose"`
	Truncated int       `xml:"truncated"`
	Difficult int       `xml:"difficult"`
	Occluded  int       `xml:"occluded"`
	Bndbox    VOCBndbox `xml:"bndbox"`
}

type VOCAnnotation struct {
	XMLName   xml.Name            `xml:"annotation"`
	Folder    string              `xml:"folder"`
	Filename  string              `xml:"filename"`
	Path      string              `xml:"path"`
	Source    VOCDataSource       `xml:"source"`
	Size      VOCImageSize        `xml:"size"`
	Segmented int                 `xml:"segmented"`
	Object    []VOCAnnotationItem `xml:"object"`
}

type VOCAnnotations []VOCAnnotation

func ReadVOCAnnotationFromFile(annotation *VOCAnnotation, path string) error {

	if pathExt := filepath.Ext(path); pathExt == "" || strings.ToLower(pathExt) != ".xml" {
		return errors.New(path + " is not a valid xml file path")
	}

	xmlBytes, readErr := ioutil.ReadFile(path)

	if readErr != nil {
		return readErr
	}

	xmlStr := string(xmlBytes)
	return xml.Unmarshal([]byte(xmlStr), annotation)
}

func ReadVOCAnnotationFromDir(annotations *VOCAnnotations, path string) error {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	isEmpty := true

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		filePath := filepath.Join(path, fileName)
		if strings.ToLower(filepath.Ext(filePath)) == ".xml" {
			var annotation VOCAnnotation
			err := ReadVOCAnnotationFromFile(&annotation, filePath)
			if err != nil {
				return err
			}
			isEmpty = false
			*annotations = append(*annotations, annotation)
		}
	}

	if isEmpty {
		return errors.New("not found xml file in the directory path")
	}

	return nil
}

func ReadVOCAnnotationsFromCOCOFile(annotations *VOCAnnotations, path string) error {

	// read the coco annotation data from json file
	var cocoAnnotations COCOAnnotations
	if err := ReadCOCOAnnotationsFromFile(&cocoAnnotations, path); err != nil {
		return err
	}

	// generate the image map with ID
	imageMap := make(map[int]COCOImage)
	annotationMap := make(map[int]VOCAnnotation)
	for _, image := range cocoAnnotations.Images {
		imageMap[image.ID] = image
		vocAnnotation := VOCAnnotation{
			Folder:   "",
			Filename: image.FileName,
			Path:     image.FileName,
			Source: VOCDataSource{
				Database:   "datasetgo.smslit.cn",
				Image:      "",
				Annotation: "",
			},
			Size: VOCImageSize{
				Width:  image.Width,
				Height: image.Height,
				Depth:  3,
			},
			Segmented: 0,
			Object:    make([]VOCAnnotationItem, 0),
		}
		annotationMap[image.ID] = vocAnnotation
	}

	// generate the category map with ID
	categoryMap := make(map[int]COCOCategory)
	for _, category := range cocoAnnotations.Categories {
		categoryMap[category.ID] = category
	}

	// make the voc annotation data from coco annotations data
	for _, annotationItem := range cocoAnnotations.Annotations {
		// check VOVAnnotation data of the image if exists
		vocAnnotation, ok := annotationMap[annotationItem.ImageID]
		if !ok {
			return fmt.Errorf("the image with ID[%v] does not exist(annotation with ID[%v])", annotationItem.ImageID, annotationItem.ID)
		}

		category, ok := categoryMap[annotationItem.CategoryID]
		if !ok {
			return fmt.Errorf("the category with ID[%v] does not exist(annotation with ID[%v])", annotationItem.CategoryID, annotationItem.ID)
		}
		// convert the coco annotation data to voc annotation data
		vocAnnotationItem := VOCAnnotationItem{
			Name:      category.Name,
			Pose:      Unspecified,
			Truncated: 0,
			Difficult: 0,
			Occluded:  0,
			Bndbox: VOCBndbox{
				Xmin: int(annotationItem.BBox[0]),
				Ymin: int(annotationItem.BBox[1]),
				Xmax: int(annotationItem.BBox[0] + annotationItem.BBox[2]),
				Ymax: int(annotationItem.BBox[1] + annotationItem.BBox[3]),
			},
		}
		vocAnnotation.Object = append(vocAnnotation.Object, vocAnnotationItem)
		annotationMap[annotationItem.ImageID] = vocAnnotation
	}

	for _, annotation := range annotationMap {
		*annotations = append(*annotations, annotation)
	}

	return nil
}

func ReadVOCAnnotationsFromCreateMLFile(annotations *VOCAnnotations, path string) error {
	var createMLAnnotations CreateMLAnnotations
	if err := ReadCreateMLAnnotationsFromFile(&createMLAnnotations, path); err != nil {
		return err
	}

	for _, createMLAnnotation := range createMLAnnotations {
		imagePath := filepath.Join(filepath.Dir(path), createMLAnnotation.Image)
		imageFile, err := os.Open(imagePath)
		if err != nil {
			return fmt.Errorf("image [%v] opening... %v", imagePath, err.Error())
		}
		defer imageFile.Close()
		imageConfig, _, err := image.DecodeConfig(imageFile)
		if err != nil {
			return fmt.Errorf("image [%v] reading... %v", imagePath, err.Error())
		}
		vocAnnotation := VOCAnnotation{
			Folder:   "",
			Filename: createMLAnnotation.Image,
			Path:     createMLAnnotation.Image,
			Source: VOCDataSource{
				Database:   "datasetgo.smslit.cn",
				Image:      "",
				Annotation: "",
			},
			Size: VOCImageSize{
				Width:  imageConfig.Width,
				Height: imageConfig.Height,
				Depth:  3,
			},
			Segmented: 0,
			Object:    make([]VOCAnnotationItem, 0),
		}

		for _, createAnnotationItem := range createMLAnnotation.Annotations {
			vocAnnotationItem := VOCAnnotationItem{
				Name:      createAnnotationItem.Label,
				Pose:      Unspecified,
				Truncated: 0,
				Difficult: 0,
				Occluded:  0,
				Bndbox: VOCBndbox{
					Xmin: int(createAnnotationItem.Coordinates.X),
					Ymin: int(createAnnotationItem.Coordinates.Y),
					Xmax: int(createAnnotationItem.Coordinates.X + createAnnotationItem.Coordinates.Width),
					Ymax: int(createAnnotationItem.Coordinates.Y + createAnnotationItem.Coordinates.Height),
				},
			}
			vocAnnotation.Object = append(vocAnnotation.Object, vocAnnotationItem)
		}

		*annotations = append(*annotations, vocAnnotation)
	}

	return nil
}

func WriteVOCAnnotationsToFile(annotations *VOCAnnotations, path string) error {
	for _, annotation := range *annotations {
		imageName := annotation.Filename
		imageExt := filepath.Ext(imageName)
		xmlName := strings.TrimSuffix(imageName, imageExt) + ".xml"
		xmlPath := filepath.Join(path, xmlName)
		if annotationBytes, err := xml.MarshalIndent(annotation, "", "    "); err != nil {
			return err
		} else {
			if writeErr := ioutil.WriteFile(xmlPath, annotationBytes, 0666); writeErr != nil {
				return writeErr
			}
		}
	}
	return nil
}
