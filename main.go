package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	md   = "markdown"
	txt  = "text"
	stop = "(STOP)"
)

type Photo struct {
	Name string
	Date time.Time
	FL   string
	ISO  string
	SS   string
	F    string
	Lens string
	Body string
}

func main() {
	imageList := fetchImageList("./")
	fmt.Println("Number of photo files: ", len(imageList))
	if len(imageList) == 0 {
		return
	}

	outputType := selectOutputType()
	if outputType == stop {
		return
	}

	if outputType == md {
		createMd(imageList)
		return
	}
	if outputType == txt {
		createTxt(imageList)
		return
	}
}

// DIR IMAGE LIST
func fetchImageList(dir string) []Photo {
	paths, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	exif.RegisterParsers(mknote.All...)

	var list []Photo
	for _, path := range paths {
		abspath := filepath.Join(dir, path.Name())
		file, _ := os.Open(abspath)
		head := make([]byte, 261)
		file.Read(head)
		if filetype.IsImage(head) {
			file, _ = os.Open(abspath)
			x, err := exif.Decode(file)
			if err != nil {
				fmt.Println("exif err:", err)
				return list
			}

			bodyModelTag, _ := x.Get(exif.Model)
			focal, _ := x.Get(exif.FocalLength)
			lensModelTag, _ := x.Get(exif.LensModel)
			ssTag, _ := x.Get(exif.ExposureTime)
			fTag, _ := x.Get(exif.FNumber)
			isoTag, _ := x.Get(exif.ISOSpeedRatings)
			date, _ := x.DateTime()
			bodyModel := strings.Replace(bodyModelTag.String(), "\"", "", -1)
			numer, _, _ := focal.Rat2(0)
			fl := strconv.FormatInt(numer, 10)
			lensModel := strings.Replace(lensModelTag.String(), "\"", "", -1)
			ss := strings.Replace(ssTag.String(), "\"", "", -1)
			fStr := strings.Replace(fTag.String(), "\"", "", -1)
			fs := strings.Split(fStr, "/")
			f1, _ := strconv.Atoi(fs[0])
			f2, _ := strconv.Atoi(fs[1])
			f := strconv.FormatFloat(float64(f1)/float64(f2), 'f', 1, 64)
			iso := isoTag.String()
			list = append(list, Photo{Name: abspath, Date: date, FL: fl, SS: ss, F: f, ISO: iso, Body: bodyModel, Lens: lensModel})
		}
	}

	return list
}

// OUTPUT FILE TYPE (select: markdown or text)
func selectOutputType() string {
	outputType := ""
	prompt := &survey.Select{
		Message: "Output type:",
		Options: []string{md, txt, stop},
	}
	survey.AskOne(prompt, &outputType, nil)

	return outputType
}

func addLine(lines []byte, item string) []byte {
	return append(lines[:], []byte(item + "\n")[:]...)
}

func createMd(imageList []Photo) {
	file, err := os.OpenFile("exiflist.md", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []byte

	// head
	lines = addLine(lines, "# exif list")
	lines = addLine(lines, "last updated: "+time.Now().String())
	lines = addLine(lines, "---")

	// body
	for _, photo := range imageList {
		lines = addLine(lines, "## "+photo.Name)
		lines = addLine(lines, "<img src='"+photo.Name+"' alt='drawing' width='200'/>")
		lines = addLine(lines, photo.Date.String())
		lines = addLine(lines, "\n")
		lines = addLine(lines, photo.Body)
		lines = addLine(lines, photo.Lens)
		lines = addLine(lines, photo.FL+"mm")
		lines = addLine(lines, photo.SS+"sec")
		lines = addLine(lines, "F"+photo.F)
		lines = addLine(lines, "ISO "+photo.ISO)
		lines = addLine(lines, "\n")
	}

	// write
	file.Write(([]byte)(lines))
}

func createTxt(imageList []Photo) {
	file, err := os.OpenFile("exiflist.txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []byte

	// head
	lines = addLine(lines, "# exif list")
	lines = addLine(lines, "last updated: "+"2019.4.11 20:20")
	lines = addLine(lines, "---")

	// body
	for _, photo := range imageList {
		lines = addLine(lines, "## "+photo.Name)
		lines = addLine(lines, photo.Date.String())
		lines = addLine(lines, "\n")
		lines = addLine(lines, photo.Body)
		lines = addLine(lines, photo.Lens)
		lines = addLine(lines, photo.FL+"mm")
		lines = addLine(lines, photo.SS+"sec")
		lines = addLine(lines, "F"+photo.F)
		lines = addLine(lines, "ISO "+photo.ISO)
		lines = addLine(lines, "\n")
	}

	// write
	file.Write(([]byte)(lines))
}
