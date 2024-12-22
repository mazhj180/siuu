package test

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

// 灰度值到字符的映射表
var chars = []rune{'@', '#', '8', '&', '%', '$', '?', '*', '+', ';', ':', ',', '.'}

func TestBanner(t *testing.T) {
	// 打开图片文件
	file, err := os.Open("/Users//Downloads/image3.jpg")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 解码图像
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	// 获取图像的宽高
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	// 按照一定比例调整图像大小
	// 可以根据需要修改图像宽度和高度的缩放比例
	scale := 0.1
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// 遍历图像的每个像素，并根据灰度值打印字符
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// 计算原图像中的对应像素位置
			originalX := int(float64(x) / scale)
			originalY := int(float64(y) / scale)

			// 获取像素的颜色
			r, g, b, _ := img.At(originalX, originalY).RGBA()

			// 计算灰度值（只取 RGB 平均值）
			gray := (r + g + b) / 3
			gray = gray >> 8 // 将范围缩小到 0-255

			// 映射灰度值到字符
			charIndex := int(gray * 10 / 255) // 映射到 0-10
			fmt.Print(string(chars[charIndex]))
		}
		fmt.Println()
	}
}
