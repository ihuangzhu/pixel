package main

import (
	"bufio"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
)

const SquarePixels = 5

const Charters = " .,;?!|-_~`@#$%^&*()=+abcdefghijklmnopqrstuvxxyzABCDEFGHIJKLMNOPQRSTUVXXYZ0123456789一二三四五六七八九十墨▇"

var ftBytes []byte
var ft *truetype.Font

var mapped map[int]image.Image
var mappedKeys []int

func init() {
	// 读取字体数据
	var err error
	ftBytes, err = ioutil.ReadFile("./fonts/msyh.ttc")
	if err != nil {
		log.Println(err)
	}
	// 载入字体数据
	ft, err = freetype.ParseFont(ftBytes)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	// 灰度素材
	mapped = make(map[int]image.Image)
	mapped[0] = image.NewAlpha(image.Rect(0, 0, SquarePixels, SquarePixels)) // 透明图片

	// 载入字符
	for _, char := range Charters {
		if gary, img := loadChar(char, SquarePixels, SquarePixels, float64(SquarePixels)); img != nil {
			mapped[gary] = img
		}
	}

	// 灰度排序
	for k := range mapped {
		mappedKeys = append(mappedKeys, k)
	}
	sort.Ints(mappedKeys)

	makePng()
	makeGif()
	fmt.Println("Done.")
}

func makeGif() {
	fileInput, err := os.Open("./images/ng.gif")
	if err != nil {
		panic(err)
	}
	defer fileInput.Close()
	img, err := gif.DecodeAll(fileInput)
	if err != nil {
		panic(err)
	}

	outGif := &gif.GIF{}
	paletted := append(palette.WebSafe, color.Transparent)
	for _, srcImg := range img.Image {
		// 创建一个新的图片
		newImg := image.NewRGBA(image.Rect(0, 0, srcImg.Bounds().Dx()*SquarePixels, srcImg.Bounds().Dy()*SquarePixels))

		for i := 0; i < srcImg.Bounds().Dx(); i++ {
			for j := 0; j < srcImg.Bounds().Dy(); j++ {
				colorRgb := srcImg.At(i, j)
				R, G, B, _ := colorRgb.RGBA()
				Y := int((0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)) * (255.0 / 65535))

				dist := 0
				last := 0
				for _, key := range mappedKeys {
					if key > Y {
						x := math.Abs(float64(key - Y))
						y := math.Abs(float64(last - Y))

						dist = key
						if x > y {
							dist = last
						}

						break
					}

					dist = key
					last = key
				}

				draw.Draw(newImg, srcImg.Bounds().Add(image.Pt(i*SquarePixels, j*SquarePixels)), mapped[dist], image.Point{}, draw.Over)
			}
		}

		palettedImage := image.NewPaletted(newImg.Bounds(), paletted)
		draw.Draw(palettedImage, newImg.Bounds(), newImg, image.Point{}, draw.Src)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 1)
	}

	// 把GIF写入文件rgb.gif
	f, err := os.Create("./images/re-ng.gif")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gif.EncodeAll(f, outGif)

	fmt.Println("Gif done.")
}

func makePng() {
	// 载入图片
	fileInput, err := os.Open("./images/cp.png")
	if err != nil {
		panic(err)
	}
	defer fileInput.Close()

	// 图片解码
	img, err := png.Decode(fileInput)
	if err != nil {
		panic(err)
	}

	// 创建一个新的图片
	newImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx()*SquarePixels, img.Bounds().Dy()*SquarePixels))

	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			colorRgb := img.At(i, j)
			R, G, B, _ := colorRgb.RGBA()
			Y := int((0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)) * (255.0 / 65535))

			dist := 0
			last := 0
			for _, key := range mappedKeys {
				if key > Y {
					x := math.Abs(float64(key - Y))
					y := math.Abs(float64(last - Y))

					dist = key
					if x > y {
						dist = last
					}

					break
				}

				dist = key
				last = key
			}

			//fmt.Println(dist)
			//garyImg := image.NewRGBA(image.Rect(0, 0, SQUARE_PIXELS, SQUARE_PIXELS))
			//draw.Draw(garyImg, garyImg.Bounds(), mapped[dist], image.Point{}, draw.Src)
			//draw.Draw(newImg, img.Bounds().Add(image.Pt(i*SQUARE_PIXELS, j*SQUARE_PIXELS)), garyImg, image.Point{}, draw.Over)
			draw.Draw(newImg, img.Bounds().Add(image.Pt(i*SquarePixels, j*SquarePixels)), mapped[dist], image.Point{}, draw.Over)
		}
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// 写入文件
	outFile, err := os.Create("./images/re-cp.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	err = png.Encode(writer, newImg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = writer.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Println("Png done.")
}

func loadChar(char rune, width, height int, ftSize float64) (int, *image.RGBA) {
	// 前景色、背景色
	fg, bg := image.Black, image.White

	// 创建图片
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), bg, image.Point{}, draw.Src)

	// 绘入文字
	fc := freetype.NewContext()
	fc.SetDPI(72)            // 设置屏幕每英寸的分辨率
	fc.SetFont(ft)           // 设置用于绘制文本的字体
	fc.SetFontSize(ftSize)   // 以磅为单位设置字体大小
	fc.SetClip(img.Bounds()) // 设置剪裁矩形以进行绘制
	fc.SetDst(img)           // 设置目标图像
	fc.SetSrc(fg)            // 设置绘制操作的源图像

	// 计算文字尺寸
	face := truetype.NewFace(ft, &truetype.Options{DPI: 72, Size: ftSize})
	bounds, advance, ok := face.GlyphBounds(char)
	if ok != true {
		log.Println("!ok")
		return 0, nil
	}

	iMaxY := int(float64(bounds.Max.Y) / 64)
	iMinY := int(float64(bounds.Min.Y) / 64)
	iHeight := iMaxY - iMinY + 1

	iWidth := int(float64(advance) / 64)

	pt := freetype.Pt((width-iWidth)/2, (height-iHeight)/2+iHeight)
	fc.DrawString(string(char), pt)

	// 计算平均数
	grayPixSum := 0
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			colorRgb := img.At(i, j)
			R, G, B, _ := colorRgb.RGBA()

			//Luma: Y = 0.2126*R + 0.7152*G + 0.0722*B
			Y := (0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)) * (255.0 / 65535)
			grayPixSum += int(Y)
		}
	}

	return grayPixSum / (width * height), img
}
