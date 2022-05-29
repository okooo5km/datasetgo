package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type CreateMLCoordinates struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"Height"`
}

type CreateMLAnnotationItem struct {
	Label       string              `json:"label"`
	Coordinates CreateMLCoordinates `json:"coordinates"`
}

type CreateMLAnnotation struct {
	Image       string                   `json:"image"`
	Annotations []CreateMLAnnotationItem `json:"annotations"`
}

type CreateMLAnnotations []CreateMLAnnotation

func ReadCreateMLAnnotationsFromFile(annotations *CreateMLAnnotations, path string) error {
	if pathExt := filepath.Ext(path); pathExt == "" || strings.ToLower(pathExt) != ".json" {
		return errors.New(path + " is not a valid json file path")
	}

	jsonBytes, readErr := ioutil.ReadFile(path)

	if readErr != nil {
		return readErr
	}

	xmlStr := string(jsonBytes)
	return json.Unmarshal([]byte(xmlStr), annotations)
}

func ReadCreateMLAnnotationsFromPascalVOCDir(annotations *CreateMLAnnotations, path string) error {
	// read the voc annotations data from the directory path
	var vocAnnotations VOCAnnotations
	if err := ReadVOCAnnotationFromDir(&vocAnnotations, path); err != nil {
		return err
	}

	for _, vocAnnotation := range vocAnnotations {
		annotationItems := []CreateMLAnnotationItem{}
		for _, vocObject := range vocAnnotation.Object {
			x := float32(vocObject.Bndbox.Xmin)
			y := float32(vocObject.Bndbox.Ymin)
			width := float32(vocObject.Bndbox.Xmax - vocObject.Bndbox.Xmin)
			height := float32(vocObject.Bndbox.Ymax - vocObject.Bndbox.Ymin)
			annotationItem := CreateMLAnnotationItem{
				Label: vocObject.Name,
				Coordinates: CreateMLCoordinates{
					X:      x,
					Y:      y,
					Width:  width,
					Height: height,
				},
			}
			annotationItems = append(annotationItems, annotationItem)
		}

		annotation := CreateMLAnnotation{
			Image:       vocAnnotation.Filename,
			Annotations: annotationItems,
		}

		*annotations = append(*annotations, annotation)
	}

	return nil
}

func ReadCreateMLAnnotationsFromCOCOFile(annotations *CreateMLAnnotations, path string) error {

	// read the coco annotation data from json file
	var cocoAnnotations COCOAnnotations
	if err := ReadCOCOAnnotationsFromFile(&cocoAnnotations, path); err != nil {
		return err
	}

	// generate the image map with ID
	imageMap := make(map[int]COCOImage)
	annotationMap := make(map[int]CreateMLAnnotation)
	for _, image := range cocoAnnotations.Images {
		imageMap[image.ID] = image
		createMLAnnotation := CreateMLAnnotation{
			Image:       image.FileName,
			Annotations: make([]CreateMLAnnotationItem, 0),
		}
		annotationMap[image.ID] = createMLAnnotation
	}

	// generate the category map with ID
	categoryMap := make(map[int]COCOCategory)
	for _, category := range cocoAnnotations.Categories {
		categoryMap[category.ID] = category
	}

	// make the voc annotation data from coco annotations data
	for _, annotationItem := range cocoAnnotations.Annotations {
		// check VOVAnnotation data of the image if exists
		creatMLAnnotation, ok := annotationMap[annotationItem.ImageID]
		if !ok {
			return fmt.Errorf("the image with ID[%v] does not exist(annotation with ID[%v])", annotationItem.ImageID, annotationItem.ID)
		}

		category, ok := categoryMap[annotationItem.CategoryID]
		if !ok {
			return fmt.Errorf("the category with ID[%v] does not exist(annotation with ID[%v])", annotationItem.CategoryID, annotationItem.ID)
		}
		// convert the coco annotation data to voc annotation data
		createMLAnnotationItem := CreateMLAnnotationItem{
			Label: category.Name,
			Coordinates: CreateMLCoordinates{
				X:      annotationItem.BBox[0],
				Y:      annotationItem.BBox[1],
				Width:  annotationItem.BBox[2],
				Height: annotationItem.BBox[3],
			},
		}
		creatMLAnnotation.Annotations = append(creatMLAnnotation.Annotations, createMLAnnotationItem)
		annotationMap[annotationItem.ImageID] = creatMLAnnotation
	}

	for _, annotation := range annotationMap {
		*annotations = append(*annotations, annotation)
	}

	return nil
}

func WriteCreateMLAnnotationsToFile(annotations *CreateMLAnnotations, path string) error {
	annotationsBytes, err := json.MarshalIndent(*annotations, "", "    ")
	if err != nil {
		return err
	}

	if writeErr := ioutil.WriteFile(path, annotationsBytes, 0666); writeErr != nil {
		return writeErr
	}
	return nil
}
