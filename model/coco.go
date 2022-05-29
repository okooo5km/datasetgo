package model

import (
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type COCOInfo struct {
	Year        string `json:"year"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Contributor string `json:"contributor"`
	URL         string `json:"url"`
	DateCreated string `json:"date_created"`
}

type COCOLicense struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

type COCOCategory struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	SuperCategory string `json:"supercategory"`
}

type COCOImage struct {
	ID           int    `json:"id"`
	License      int    `json:"license"`
	FileName     string `json:"file_name"`
	Height       int    `json:"height"`
	Width        int    `json:"width"`
	DateCaptured string `json:"date_captured"`
}

type COCOAnnotation struct {
	ID           int       `json:"id"`
	ImageID      int       `json:"image_id"`
	CategoryID   int       `json:"category_id"`
	BBox         []float32 `json:"bbox"`
	Area         float32   `json:"area"`
	Segmentation []float32 `json:"segmentation"`
	IsCrowd      int       `json:"iscrowd"`
}

type COCOAnnotations struct {
	Info        COCOInfo         `json:"info"`
	Licenses    []COCOLicense    `json:"licenses"`
	Categories  []COCOCategory   `json:"categories"`
	Images      []COCOImage      `json:"images"`
	Annotations []COCOAnnotation `json:"annotations"`
}

func ReadCOCOAnnotationsFromFile(annotations *COCOAnnotations, path string) error {
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

func ReadCOCOAnnotationsFromPascalVOCDir(annotations *COCOAnnotations, path string) error {
	// read the voc annotations data from the directory path
	var vocAnnotations VOCAnnotations
	if err := ReadVOCAnnotationFromDir(&vocAnnotations, path); err != nil {
		return err
	}

	info := COCOInfo{
		Year:        "2022",
		Version:     "1",
		Description: "Exported from datasetgo",
		Contributor: "5km@smslit.cn",
		URL:         "",
		DateCreated: time.Now().Format("2006-01-02T15:04:05+00:00"),
	}

	license := COCOLicense{
		ID:   1,
		URL:  "",
		Name: "5km",
	}

	categoriesMap := make(map[string]COCOCategory)

	images := make([]COCOImage, len(vocAnnotations))

	var annotationItems []COCOAnnotation

	// add new annotation data
	for index, vocAnnotation := range vocAnnotations {

		// add new image info
		// TODO: get the captured time of image
		cocoImage := COCOImage{
			ID:           index + 1,
			License:      1,
			FileName:     vocAnnotation.Filename,
			Height:       vocAnnotation.Size.Height,
			Width:        vocAnnotation.Size.Width,
			DateCaptured: "",
		}
		images[index] = cocoImage

		// add new annotation info
		annotaionsLength := len(annotationItems)
		for i, obj := range vocAnnotation.Object {
			// get id of the category, or add new category
			category, ok := categoriesMap[obj.Name]
			if !ok {
				mapLength := len(categoriesMap)
				category = COCOCategory{
					ID:            mapLength + 1,
					Name:          obj.Name,
					SuperCategory: "",
				}
				categoriesMap[obj.Name] = category
			}
			boxWidth := float32(obj.Bndbox.Xmax - obj.Bndbox.Xmin)
			boxHeight := float32(obj.Bndbox.Ymax - obj.Bndbox.Ymin)
			annotationItem := COCOAnnotation{
				ID:           annotaionsLength + i,
				ImageID:      cocoImage.ID,
				CategoryID:   category.ID,
				BBox:         []float32{float32(obj.Bndbox.Xmin), float32(obj.Bndbox.Ymin), boxWidth, boxHeight},
				Area:         boxWidth * boxHeight,
				Segmentation: make([]float32, 0),
				IsCrowd:      0,
			}
			annotationItems = append(annotationItems, annotationItem)
		}

	}

	// generate category data from map data
	categories := make([]COCOCategory, len(categoriesMap))
	for _, category := range categoriesMap {
		categories[category.ID-1] = category
	}

	*annotations = COCOAnnotations{
		Info:        info,
		Licenses:    []COCOLicense{license},
		Images:      images,
		Categories:  categories,
		Annotations: annotationItems,
	}

	return nil
}

func ReadCOCOAnnotationsFromCreateMLFile(annotations *COCOAnnotations, path string) error {
	var createMLAnnotations CreateMLAnnotations
	if err := ReadCreateMLAnnotationsFromFile(&createMLAnnotations, path); err != nil {
		return err
	}

	info := COCOInfo{
		Year:        "2022",
		Version:     "1",
		Description: "Exported from datasetgo",
		Contributor: "5km@smslit.cn",
		URL:         "",
		DateCreated: time.Now().Format("2006-01-02T15:04:05+00:00"),
	}

	license := COCOLicense{
		ID:   1,
		URL:  "",
		Name: "5km",
	}

	categoriesMap := make(map[string]COCOCategory)

	images := make([]COCOImage, len(createMLAnnotations))

	var annotationItems []COCOAnnotation

	// add new annotation data
	for index, createMLAnnotation := range createMLAnnotations {
		imagePath := filepath.Join(filepath.Dir(path), createMLAnnotation.Image)
		imageFile, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		defer imageFile.Close()
		imageConfig, _, err := image.DecodeConfig(imageFile)
		if err != nil {
			return err
		}
		// add new image info
		// TODO: get the captured time of image
		cocoImage := COCOImage{
			ID:           index + 1,
			License:      1,
			FileName:     createMLAnnotation.Image,
			Height:       imageConfig.Height,
			Width:        imageConfig.Width,
			DateCaptured: "",
		}
		images[index] = cocoImage

		// add new annotation info
		annotaionsLength := len(annotationItems)
		for i, createMLAnnotationItem := range createMLAnnotation.Annotations {
			// get id of the category, or add new category
			category, ok := categoriesMap[createMLAnnotationItem.Label]
			if !ok {
				mapLength := len(categoriesMap)
				category = COCOCategory{
					ID:            mapLength + 1,
					Name:          createMLAnnotationItem.Label,
					SuperCategory: "",
				}
				categoriesMap[createMLAnnotationItem.Label] = category
			}
			annotationItem := COCOAnnotation{
				ID:           annotaionsLength + i,
				ImageID:      cocoImage.ID,
				CategoryID:   category.ID,
				BBox:         []float32{createMLAnnotationItem.Coordinates.X, createMLAnnotationItem.Coordinates.Y, createMLAnnotationItem.Coordinates.Width, createMLAnnotationItem.Coordinates.Height},
				Area:         createMLAnnotationItem.Coordinates.Width * createMLAnnotationItem.Coordinates.Height,
				Segmentation: make([]float32, 0),
				IsCrowd:      0,
			}
			annotationItems = append(annotationItems, annotationItem)
		}

	}

	// generate category data from map data
	categories := make([]COCOCategory, len(categoriesMap))
	for _, category := range categoriesMap {
		categories[category.ID-1] = category
	}

	*annotations = COCOAnnotations{
		Info:        info,
		Licenses:    []COCOLicense{license},
		Images:      images,
		Categories:  categories,
		Annotations: annotationItems,
	}

	return nil
}

func WriteCOCOAnnotationsToFile(annotations *COCOAnnotations, path string) error {
	annotationsBytes, err := json.MarshalIndent(*annotations, "", "    ")
	if err != nil {
		return err
	}

	if writeErr := ioutil.WriteFile(path, annotationsBytes, 0666); writeErr != nil {
		return writeErr
	}
	return nil
}
