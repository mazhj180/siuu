package test

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

//go:embed ronaldo.txt
var images []byte

func TestAscii(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewReader(images))

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

// 灰度值到字符的映射表
var chars = []rune("@%#*+=-:. ")

func TestBanner(t *testing.T) {
	// 打开图片文件
	file, err := os.Open("/Users/mazhj/Downloads/使用工具生成图片.png")
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
	newHeight := int(float64(height) * scale * 0.5)

	// 遍历图像的每个像素，并根据灰度值打印字符
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// 计算原图像中的对应像素位置
			originalX := int(float64(x) / scale)
			originalY := int(float64(y) / scale * 0.5)

			if originalX >= width {
				originalX = width - 1
			}
			if originalY >= height {
				originalY = height - 1
			}

			// 获取像素的颜色
			r, g, b, _ := img.At(originalX, originalY).RGBA()

			// 计算灰度值（加权平均，更符合人眼感知）
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			gray = gray / 65535.0 // 将 RGBA 值转换到 0-1 范围

			// 映射灰度值到字符
			charIndex := int(gray * float64(len(chars)-1))
			if charIndex < 0 {
				charIndex = 0
			} else if charIndex >= len(chars) {
				charIndex = len(chars) - 1
			}
			// 在字符之间添加空格以增加间距
			fmt.Print(string(chars[charIndex]) + " ")
		}
		fmt.Println()
	}
}
