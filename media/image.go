package media

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/gift"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
	"time"
)

/*
 * Image resize interface
 */

//inter macro define
const (
	ImgResize = "resize"
	ImgCropToSize = "cropToSize"
	ImgSigmoid = "sigmoid"
)

//inter default value
const (
	ImgDefaultScaleWidth = 240
	ImgDefaultCropHeight = 120

	AvatarWidth = 64
	AvatarHeight = 64

	DefaultTempPath = "./"
)

//image info
type ImageResize struct {
	filters []string `gift filters`
	gfs map[string]*gift.GIFT `gift instances`
	scaleWidth int `image scale max width`
	cropHeight int `image crop max height`
	needCrop bool `image crop or not`
	tempPath string `image temp path`
}

//construct
func NewImageResize(scaleWidths ...int) *ImageResize {
	scaleWidth := ImgDefaultScaleWidth
	if scaleWidths != nil {
		scaleWidth = scaleWidths[0]
	}
	this := NewImageResizeWithPara(
			scaleWidth,
			ImgDefaultCropHeight,
			DefaultTempPath,
		)
	return this
}
func NewImageResizeWithPara(
			scaleWidth, cropHeight int,
			tempPath string,
			needCrops ...bool,
		) *ImageResize {
	var (
		filters = make(map[string]gift.Filter)
	)

	//check and set default value
	if scaleWidth <= 0 {
		scaleWidth = ImgDefaultScaleWidth
	}
	if cropHeight <= 0 {
		cropHeight = ImgDefaultCropHeight
	}

	needCrop := false
	if needCrops != nil {
		needCrop = needCrops[0]
	}

	//init batch filters
	filters[ImgResize] = gift.Resize(scaleWidth, 0, gift.LanczosResampling)
	if needCrop {
		filters[ImgCropToSize] = gift.CropToSize(scaleWidth, cropHeight, gift.CenterAnchor)
	}

	//self init
	this := &ImageResize{
		filters:make([]string, 0),
		gfs:make(map[string]*gift.GIFT),
		scaleWidth:scaleWidth,
		cropHeight:cropHeight,
		needCrop:needCrop,
		tempPath:tempPath,
	}

	//init batch gift
	for name, filter := range filters {
		this.filters = append(this.filters, name)
		this.gfs[name] = gift.New(filter)
	}

	return this
}

// Decode reads and analyzes the given reader as a GIF image
func (i *ImageResize) SplitAnimatedGif(
			fileName string, needAll bool,
		) (bool, []string) {
	var (
		frameFile, frameFilePath string
		frameFiles = make([]string, 0)
	)

	if fileName == "" {
		return false, nil
	}

	defer func() {
		if err := recover(); err != nil {
			log.Println("ImageResize::SplitAnimatedGif`rtip、failed, err:", err)
		}
	}()

	//try open file
	filePath := fmt.Sprintf("%s/%s", i.tempPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return false, nil
	}

	//delay close file
	defer file.Close()

	//decode
	gif, err := gif.DecodeAll(file)
	if err != nil {
		log.Println("ImageResize::SplitAnimatedGif failed, err:", err.Error())
		return false, nil
	}

	//frames
	frames := len(gif.Image)

	//get gif width and height
	imgWidth, imgHeight := i.getGifDimensions(gif)
	overPaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overPaintImage, overPaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	//get now time
	now := time.Now().Unix()

	if !needAll {
		randomFrameIdx := int(time.Now().Unix()) % frames
		frameFile = fmt.Sprintf("%d_%d.png", now, randomFrameIdx)
		frameFilePath = fmt.Sprintf("%s/%s", i.tempPath, frameFile)

		srcImg := gif.Image[randomFrameIdx]
		draw.Draw(overPaintImage, overPaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		//save frame image
		i.SaveImage(frameFilePath, overPaintImage)

		//add into result
		frameFiles = append(frameFiles, frameFile)
		return true, frameFiles
	}

	//process all frames
	for x, srcImg := range gif.Image {
		draw.Draw(overPaintImage, overPaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		frameFile = fmt.Sprintf("%d_%d.png", now, x)
		frameFilePath = fmt.Sprintf("%s/%s", i.tempPath, frameFile)
		i.SaveImage(frameFilePath, overPaintImage)
		//add into result
		frameFiles = append(frameFiles, frameFile)
	}

	return true, frameFiles
}

//resize image from file
func (i *ImageResize) ResizeFromFullFile(fileSrc, fileDst string) error {
	//check
	if fileSrc == "" || fileDst == "" {
		return errors.New("invalid parameter")
	}

	//load src image
	src, err := i.LoadImage(fileSrc)
	if err != nil {
		return err
	}
	if src == nil {
		return errors.New("load image file failed")
	}

	//process image by filters
	for _, name := range i.filters {
		gf, ok := i.gfs[name]
		if !ok {
			continue
		}
		src = i.ProcessOneGift(gf, src)

		//begin save target image
		err = i.SaveImage(fileDst, src)
	}
	return err
}

//resize image from file
//return targetFileName, error
func (i *ImageResize) ResizeFromFile(fileName string) (string, error) {
	var (
		finalFileName string
		//dst *image.Image
	)
	if fileName == "" {
		return finalFileName, errors.New("invalid parameter")
	}

	//format temp file path
	filePath := fmt.Sprintf("%s/%s", i.tempPath, fileName)

	//load original image
	src, err := i.LoadImage(filePath)
	if err != nil || src == nil {
		return finalFileName, err
	}

	////reduce height
	//size := src.Bounds().Size()
	//log.Println("size:", size)
	//reSized, err := carve.ReduceHeight(src, 50)
	//if err != nil {
	//	return false, finalFileName
	//}
	//fileName := path.Base(filePath)
	//log.Println("ImageService::ResizeFromFile, tempFile:", tempFile)

	//begin resize and crop image
	for _, name := range i.filters {
		gf, ok := i.gfs[name]
		if !ok {
			continue
		}
		//dst = image.NewNRGBA(gf.Bounds(src.Bounds()))
		//gf.Draw(dst, src)
		src = i.ProcessOneGift(gf, src)
	}

	//dst := image.NewNRGBA(i.gf.Bounds(src.Bounds()))
	//i.gf.Draw(dst, src)

	//generate final file name
	//fileName := path.Base(filePath)
	finalFileName = fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)

	//begin save dst image
	finalTempFile := fmt.Sprintf("%s/%s", i.tempPath, finalFileName)
	err = i.SaveImage(finalTempFile, src)
	return finalFileName, err
}

//process one gift
func (i *ImageResize) ProcessOneGift(gf *gift.GIFT, src image.Image) image.Image {
	dst := image.NewNRGBA(gf.Bounds(src.Bounds()))
	gf.Draw(dst, src)
	return dst
}

//resize image from io reader
//convert to []byte, error
func (i *ImageResize) ResizeFromIOReader(
				reader *os.File,
				isPng ...bool,
			) ([]byte, error) {
	//check
	if reader == nil {
		return nil, errors.New("invalid file")
	}
	//try decode image
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	//begin resize and crop image
	for _, name := range i.filters {
		gf, ok := i.gfs[name]
		if !ok {
			continue
		}
		img = i.ProcessOneGift(gf, img)
	}

	//convert to bytes
	buf := new(bytes.Buffer)
	if isPng != nil && isPng[0] {
		//for png
		err = png.Encode(buf, img)
	}else{
		//for jpg
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//save dst image file
func (i *ImageResize) SaveImage(filePath string, img image.Image) error {
	//check
	if filePath == "" || img == nil {
		return errors.New("invalid parameter")
	}

	//try create file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	//save image data
	defer f.Close()

	//format file info
	tempSlice := strings.Split(filePath, ".")
	tempLen := len(tempSlice)
	kind := tempSlice[tempLen-1]

	//encode image
	switch kind {
	case "jpg":
		{
			err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		}
	default:
		{
			err = png.Encode(f, img)
		}
	}
	return err
}

//load file
func (i *ImageResize) LoadFile(filePath string) (*os.File, error) {
	//try open file
	f, err := os.Open(filePath)
	return f, err
}

//load original image file
func (i *ImageResize) LoadImage(filePath string) (image.Image, error) {
	//try open file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	//try decode image
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

//get gif dimension
//return x, y width
func (i *ImageResize) getGifDimensions(gif *gif.GIF) (int, int) {
	var (
		lowestX int
		lowestY int
		highestX int
		highestY int
	)
	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}
	return highestX - lowestX, highestY - lowestY
}