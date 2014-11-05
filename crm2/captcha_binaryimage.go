package crm2

import (
	"bytes"
	"fmt"
	"image"
	"strconv"
)

// 脚的 定义成内嵌 [][]int 的struct更方便,懒得改了
type BinaryImage [][]int

// !panic
func (bi BinaryImage) String() string {
	binimg := [][]int(bi)
	h := len(binimg)
	if h < 1 {
		panic("len(binimg) < 1")
	}
	w := len(binimg[0])
	if w < 1 {
		panic("len(binimg[0] < 1")
	}
	return bi.RectString(image.Rect(0, 0, w-1, h-1))
}

func (bi BinaryImage) RectString(rect image.Rectangle) string {
	binimg := [][]int(bi)
	buf := bytes.NewBuffer(nil)
	for y, maxy := rect.Min.Y, rect.Max.Y; y <= maxy; y++ {
		for x, maxx := rect.Min.X, rect.Max.X; x <= maxx; x++ {
			/*_, err := */ buf.WriteString(strconv.Itoa(binimg[y][x]))
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

// 返回sub BinaryImage
func (bi BinaryImage) SubBinaryImage(rect image.Rectangle) BinaryImage {
	binimg := BinaryImage(bi)
	h, w := rect.Size().Y+1, rect.Size().X+1

	// init
	newbi := make([][]int, h, h)
	for y := 0; y < h; y++ {
		newbi[y] = make([]int, w, w)
	}

	// copy
	for y := 0; y < h; y++ {
		copy(newbi[y], binimg[rect.Min.Y+y][rect.Min.X:rect.Max.X+1])
	}
	return BinaryImage(newbi)
}

// !panic 数组越界 返回切割成n块的区域
func (bi BinaryImage) CropRect(n int) []image.Rectangle {
	rectes := make([]image.Rectangle, n, n)
	binimg := [][]int(bi)
	maxY := len(binimg)
	if maxY == 0 {
		panic("empty binaryimage : h == 0")
	}
	maxX := len(binimg[0])
	if maxX == 0 {
		panic("empty binaryimage : w == 0")
	}

	XS := make([]bool, maxX, maxX)
	for x := 0; x < maxX; x++ {
		for y := 0; y < maxY; y++ {
			if binimg[y][x] == 1 {
				XS[x] = true
				break
			}
		}
	}

	minXs := make([]int, 0, 6)
	maxXs := make([]int, 0, 6)
	// 验证码可能在最后竖列,(暂时未发现可能在第一竖列)
	// bug 必须连续两竖排才OK --> 小写I 不对
	// bug 不能处理不连续的...
	for x := 1; x < maxX-1; x++ {
		switch {
		//case : // 不连续,暂时不考虑边缘数组越界
		//
		//case : // 不连续,暂时不考虑边缘数组越界
		//

		case !XS[x-1] && XS[x] && XS[x+1]: // 连续
			minXs = append(minXs, x-1)
		case XS[x-1] && XS[x] && !XS[x+1]: // 连续
			maxXs = append(maxXs, x+1)
		case !XS[x-1] && XS[x] && !XS[x+1]: // 小写I
			count := 0
			for y := 0; y < maxY; y++ {
				// 中间像素>10,两旁无像素
				if binimg[y][x-1] == 1 || binimg[y][x+1] == 1 {
					goto end
				}
				if binimg[y][x] == 1 { // black
					count++
				}
			}
			if count >= 10 {
				minXs = append(minXs, x-1)
				maxXs = append(maxXs, x+1)
			}
		end:
		}
	}

	// 最后一竖列
	// todo 最后也必须验证三个
	if len(maxXs) == 5 && XS[maxX-2] && XS[maxX-1] {
		maxXs = append(maxXs, maxX-1)
	}

	if len(maxXs) != 6 || len(minXs) != 6 {
		panic(fmt.Sprintf("len(maxXs) = %d || len(minXs) = %d", len(maxXs), len(minXs)))
	}

	for i := 0; i < 6; i++ {
		rectes[i] = image.Rectangle{image.Point{minXs[i], 0}, image.Point{maxXs[i], 0}}
	}

	// 针对每个x区域扫描y
	YS := make([]bool, maxY, maxY)
	for i := 0; i < 6; i++ {

		for y := 0; y < maxY; y++ {
			for x := rectes[i].Min.X; x < rectes[i].Max.X; x++ {
				if binimg[y][x] == 1 {
					YS[y] = true
					break
				}
			}
		}

		// bug : 未考虑最后一个
		// 验证码可能在第一横行和最后一横行,所以必须考虑特殊情况
		// 扩展1像素 ??
		// 验证三个 小写I 不对

		for y, l := 1, maxY>>1; y < l; y++ {
			if !YS[y-1] && YS[y] && YS[y+1] { // 连续
				rectes[i].Min.Y = y - 1
				break
			}
		}
		for y, l := maxY-2, maxY>>1; y > l; y-- {
			if YS[y-1] && YS[y] && !YS[y+1] { // 连续
				rectes[i].Max.Y = y + 1
				break
			}
		}
		// 边缘
		if YS[maxY-2] && YS[maxY-1] {
			rectes[i].Max.Y = maxY - 1
		}

		// YS Reset
		for j := 0; j < maxY; j++ {
			YS[j] = false
		}
	}
	return rectes
}

// !panic 返回n块复制的切割区域
func (bi BinaryImage) CropSubImg(n int) []BinaryImage {
	rectes := bi.CropRect(n)
	subbis := make([]BinaryImage, n, n)
	for i := 0; i < n; i++ {
		subbis[i] = bi.SubBinaryImage(rectes[i])
	}
	return subbis
}

func (bi BinaryImage) CropSubImgNoPanic(n int) (subbis []BinaryImage) {
	subbis = make([]BinaryImage, n, n)
	defer func() {
		if recover() != nil {
			subbis = nil
		}
	}()
	return bi.CropSubImg(n)
}

// !panic(未检测BinaryImage是否为空) 计算相似度 <= 5 相似 大于 10 不同
func (bi BinaryImage) Similarity(anobi BinaryImage) int {
	a, b := [][]int(bi), [][]int(anobi)
	var h, w int
	if ha, hb := len(a), len(b); ha < hb {
		h = ha
	} else {
		h = hb
	}
	if wa, wb := len(a[0]), len(b[0]); wa < wb {
		w = wa
	} else {
		w = wb
	}
	pfa := bi.FingerPrint(h, w)
	pfb := anobi.FingerPrint(h, w)

	return Hamming(pfa, pfb)
}

// !panic(未检测BinaryImage是否为空) 计算指纹
func (bi BinaryImage) FingerPrint(h, w int) []byte {
	binimg := [][]int(bi)
	var (
		sum float32
		avg float32
	)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum += float32(binimg[y][x])
		}
	}

	avg = sum / (float32(h) * float32(w))

	buf := bytes.NewBuffer(nil)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if float32(binimg[y][x]) >= avg {
				buf.WriteByte('1')
			} else {
				buf.WriteByte('0')
			}
		}
	}
	return buf.Bytes()
}
