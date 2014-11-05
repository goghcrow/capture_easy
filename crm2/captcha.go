package crm2

import (
	"bufio"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type Alpha byte

// 仅支持大小写字母与数字
type Captcha struct {
	N         int //验证码图片字符数量
	Threshold int // 阈值 34000
	StdModule map[Alpha]BinaryImage
}

// constructor
func NewCaptcha( /*image image.Image,*/ threshold, n int) *Captcha {
	return &Captcha{
		N: n,
		//Image:     image,
		Threshold: threshold,
	}
}

// 通用二值化
func (c *Captcha) Binarify(img image.Image) BinaryImage {
	// 去除1px边缘
	h := img.Bounds().Dy() - 2
	w := img.Bounds().Dx() - 2

	//int Img[h][w]
	binimg := make([][]int, h)
	for i, l := 0, h; i < l; i++ {
		binimg[i] = make([]int, w)
	}

	for y := 1; y < h+1; y++ {
		for x := 1; x < w+1; x++ {
			// ps里头查看验证码,发现蓝通道对比较大
			// 直接通过蓝通道二值化处理
			_, _, b, _ := img.At(x, y).RGBA()
			if b < uint32(c.Threshold) {
				binimg[y-1][x-1] = 1
			} else {
				binimg[y-1][x-1] = 0
			}
		}
	}
	return BinaryImage(binimg)
}

// !panic 二值化并 切割并复制子区域
func (c *Captcha) Crop(img image.Image) []BinaryImage {
	return c.Binarify(img).CropSubImgNoPanic(c.N)
}

// 从文件读取并解码模板
func (c *Captcha) decodeModuleFile(moduleFile string, to interface{}) error {
	f, err := os.Open(moduleFile)
	if err != nil {
		return err
	}
	defer f.Close()
	en, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	err = ByteDecode(en, to)
	if err != nil {
		return err
	}
	return nil
}

// 将模板写入文件 文件存在则覆盖
func (c *Captcha) encodeModuleFile(module interface{}, moduleFile string) error {
	f, err := os.Create(moduleFile)
	if err != nil {
		return err
	}
	defer f.Close()

	en, err := ByteEncode(module)
	if err != nil {
		return err
	}

	_, err = f.Write(en)
	if err != nil {
		return err
	}
	return nil
}

// 载入标准模板文件 (notice:自动调用ImportStdModule,覆盖就stdMobule)
func (c *Captcha) LoadStdModule(stdfile string) (map[Alpha]BinaryImage, error) {
	stdModule := make(map[Alpha]BinaryImage)
	err := c.decodeModuleFile(stdfile, &stdModule)
	if err != nil {
		return nil, err
	}
	c.ImportStdModule(stdModule)
	return stdModule, nil
}

// 更新标准模板某个字符 Load -> update 未写入文件 -> Save
func (c *Captcha) UpdateStdModule(a Alpha, binimg BinaryImage) {
	if c.StdModule == nil {
		panic("must improtStdModule first")
	}
	c.StdModule[a] = binimg
}

// 保存标准模板文件
func (c *Captcha) SaveStdModule(stdModule map[Alpha]BinaryImage, stdfile string) error {
	return c.encodeModuleFile(stdModule, stdfile)
}

// 载入训练文件
func (c *Captcha) LoadTrainModule(trainFile string) (map[Alpha][]BinaryImage, error) {
	trainModule := make(map[Alpha][]BinaryImage)
	err := c.decodeModuleFile(trainFile, &trainModule)
	if err != nil {
		return nil, err
	}
	return trainModule, nil
}

// 保存训练文件
func (c *Captcha) SaveTrainModule(trainModule map[Alpha][]BinaryImage, trainFile string) error {
	return c.encodeModuleFile(trainModule, trainFile)
}

// 手动从训练模板生成标准模板 未写入文件 需手动调用 SaveStdModule
func (c *Captcha) ManualGenStdModuleFromFile(trainFileFrom string) (map[Alpha]BinaryImage, error) {
	trainModule, err := c.LoadTrainModule(trainFileFrom)
	if err != nil {
		return nil, err
	}
	stdModule := make(map[Alpha]BinaryImage)

	for alpha, binimges := range trainModule {
		stdModule[alpha] = binimges[c.manualFindMatchest(alpha, binimges)]
	}
	return stdModule, nil
}

// 手动寻找最匹配
func (c *Captcha) manualFindMatchest(alpha Alpha, binimges []BinaryImage) int {
	for i, l := 0, len(binimges); i < l; i++ {
		fmt.Printf("==================== Index :[ %d ] ==============\n", i)
		fmt.Printf("==================== Alpha :[ %s ] ==============\n", string(byte(alpha)))
		fmt.Println(binimges[i])
	}

	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadSlice('\n')
	if err != nil {
		log.Fatal(err)
	}

	index, err := strconv.Atoi(string(line[0]))
	if err != nil {
		log.Fatal(err)
	}
	return index
}

// 自动从训练模板生成标准模板 未写入文件 需手动调用 SaveStdModule
func (c *Captcha) AutoGenStdModuleFromFile(trainFileFrom string) (map[Alpha]BinaryImage, error) {
	trainModule, err := c.LoadTrainModule(trainFileFrom)
	if err != nil {
		return nil, err
	}
	return c.AutoGenStdModuleFromMemory(trainModule)
}

// 自动从训练模板生成标准模板 未写入文件 需手动调用 SaveStdModule
func (c *Captcha) AutoGenStdModuleFromMemory(trainModule map[Alpha][]BinaryImage) (map[Alpha]BinaryImage, error) {
	stdModule := make(map[Alpha]BinaryImage)

	for alpha, binimges := range trainModule {
		stdModule[alpha] = binimges[c.autoFindMatchest(binimges)]
	}
	return stdModule, nil
}

// 自动寻找最匹配
func (c *Captcha) autoFindMatchest(binimges []BinaryImage) int {
	// 规则,比较[]BinImg指纹位数
	// 以占数量最多的位数为准
	// 在同样位数的[]BinImg寻找每一位出现0与1的数量,以次数多的为准
	// 最终生成平均的指纹...覆盖大部分情况...

	bl := len(binimges)

	/////////////////////////
	// 生成标准化指纹,取数量最多者
	// 指纹序列组
	fpseq := make([][]byte, bl, bl)
	// 指纹序列长度频数序列
	lenseq := make(map[int]int) // 这里如果使用俩slice代替map,生成norLseq就可以生一次遍历len()
	for i := 0; i < bl; i++ {
		h, w := len(binimges[i]), len(binimges[i][0]) // !panic empty bi
		fpseq[i] = binimges[i].FingerPrint(h, w)
		tl := len(fpseq[i])
		if _, ok := lenseq[tl]; ok {
			lenseq[tl]++
		} else {
			lenseq[tl] = 1
		}
	}

	// 去频数序列中最大值
	norL := 0 // 指纹标准长度值
	maxC := 0 // 指纹标准长度值在序列中个数
	for l, c := range lenseq {
		if c > maxC {
			maxC = c
			norL = l
		}
	}

	// 从指纹序列组中去除小于长度频数的指纹序列
	norFpseq := make([][]byte, 0, maxC) // 标准化之后的指纹序列组
	//norFpseq := make([][]byte, maxC, maxC) // 标准化之后的指纹序列组
	// init
	//for i := 0; i < maxC; i++ {
	//	norFpseq[i] = make([]byte, norL, norL)
	//}
	//currI := 0
	for i := 0; i < bl; i++ {
		if len(fpseq[i]) == norL {
			norFpseq = append(norFpseq, fpseq[i])
			//norFpseq[currI] = fpseq[i]
			//currI++
		}
	}

	// 标准指纹序列
	norFp := make([]byte, norL, norL)
	for i := 0; i < norL; i++ { // 序列中第几位
		oneC := 0
		zeroC := 0
		for j := 0; j < maxC; j++ { // j 序列组中第几个序列
			if norFpseq[j][i] == 1 {
				oneC++
			} else {
				zeroC++
			}
		}
		// 宁多勿少
		if oneC >= zeroC {
			norFp[i] = 0
		} else {
			norFp[i] = 1
		}
	}
	////////////////////////////

	//  取[]BinaryImage与标准化指纹最similar值最小的index
	minN := 100
	minIndex := 0
	for i := 0; i < bl; i++ {
		if len(fpseq[i]) == norL {
			if n := Hamming(fpseq[i], norFp); n < minN {
				minN = n
				minIndex = i
			}
		}
	}
	return minIndex
}

// 从已有文件导入训练或者重新(trainFile = nil)训练模板,返回训练模板 未写入文件 需手动调用 SaveTrainModule
func (c *Captcha) Train(capimgs []image.Image, trainFile interface{}) (map[Alpha][]BinaryImage, error) {
	trainModule := make(map[Alpha][]BinaryImage)
	if trainFile != nil {
		var err error
		trainModule, err = c.LoadTrainModule(trainFile.(string))
		if err != nil {
			return nil, err
		}
	}

	r := bufio.NewReader(os.Stdin)
	for i, l := 0, len(capimgs); i < l; i++ {

		capbinimg := c.Binarify(capimgs[i])

		fmt.Println(capbinimg)
		fmt.Println("enter captcha: ")

	input:
		captch, err := r.ReadSlice('\n')
		//fmt.Println(len(captch)) // 8 = 6+ \n \ r
		if len(captch) != (c.N + 2) {
			fmt.Println("len(captch) != c.N")
			goto input
		}
		if err != nil {
			fmt.Println(err)
			goto input
		}
		binimges := capbinimg.CropSubImgNoPanic(c.N)

		if binimges == nil {
			fmt.Println("crop failed,next")
			continue
		}
		// 将输入的验证码与crop关联存入训练module
		for n := 0; n < c.N; n++ {
			alpha := Alpha(captch[n])
			if _, ok := trainModule[alpha]; ok {
				trainModule[alpha] = append(trainModule[alpha], binimges[n])
			} else {
				trainModule[alpha] = []BinaryImage{binimges[n]}
			}
		}

	}
	return trainModule, nil
}

// 标准模板检查 测试训练后缺少哪些字符
func (c *Captcha) StdModuleCheck(stdModule map[Alpha]BinaryImage, show bool) []byte {
	lacks := make([]byte, 0, 62)

	var char byte
	if show {
		fmt.Println("lack : ")
		fmt.Print("    num : ")
	}
	for char = '0'; char <= '9'; char++ {
		if _, ok := stdModule[Alpha(char)]; !ok {
			if show {
				fmt.Printf("(last %s)current [%s] ", string(char-1), string(char))
			}
			lacks = append(lacks, char)
		}
	}
	if show {
		fmt.Print("\n    upper : ")
	}
	for char = 'A'; char <= 'Z'; char++ {
		if _, ok := stdModule[Alpha(char)]; !ok {
			if show {
				fmt.Printf("(last %s)current [%s] ", string(char-1), string(char))
			}
			lacks = append(lacks, char)
		}
	}
	if show {
		fmt.Println("\n    lower : ")
	}
	for char = 'a'; char <= 'z'; char++ {
		if _, ok := stdModule[Alpha(char)]; !ok {
			if show {
				fmt.Printf("(last %s)current [%s] ", string(char-1), string(char))
			}
			lacks = append(lacks, char)
		}
	}
	if show {
		fmt.Println()
	}
	return lacks
}

// 导入标准模板,再次导入执行覆盖
func (c *Captcha) ImportStdModule(stdModule map[Alpha]BinaryImage) {
	if c.StdModule == nil {
		c.StdModule = make(map[Alpha]BinaryImage)
	}
	c.StdModule = stdModule
}

// 识别前必须先导入标准模板
func (c *Captcha) Recognize(img image.Image) string {
	if c.StdModule == nil {
		panic("must improtStdModule first")
	}

	cropsBinimgs := c.Binarify(img).CropSubImg(c.N)

	var recognized = make([]byte, c.N, c.N)

	for i := 0; i < c.N; i++ {

		var (
			min int = 100 // 汉明距离
			A   Alpha
		)

		for alpha, stdbinimg := range c.StdModule {
			hamming := stdbinimg.Similarity(cropsBinimgs[i])
			if hamming < min {
				min = hamming
				A = alpha
			}
		}

		recognized[i] = byte(A)
	}
	return string(recognized)
}
