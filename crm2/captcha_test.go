package crm2

import (
	"fmt"
	"image"
	"lib/dbgutil"
	"log"
	"testing"
)

const (
	Threshole = 34000
	N         = 6
)

var c *Captcha
var img image.Image

func init() {
	var err error
	img, err = ReadImg(`d:\CCHelper\Golang\bin\验证码\0.jpg`)
	if err != nil {
		log.Fatal(err)
	}
	c = NewCaptcha(Threshole, N)
}

// 只简单测试了函数的功能,嵌套函数只测试最外层函数
// 加下划线(暂时注释):通过测试...每个函数单独测试

/// test binaryimage Region

// OK
func TestBiString(t *testing.T) {
	t.Log(c.Binarify(img))
	crops := c.Crop(img)
	for i, l := 0, len(crops); i < l; i++ {
		t.Log(crops[i])
	}
}

// OK
func TestSimilarity(t *testing.T) {
	crops := c.Crop(img)
	sameN := crops[0].Similarity(crops[0])
	diffN := crops[0].Similarity(crops[1])
	t.Log(sameN)
	t.Log(diffN)
	if sameN != 0 {
		t.Fail()
	}
	if diffN <= 5 {
		t.Fail()
	}

}

/// test capture Region

// OK
func TestBinarify(t *testing.T) {
	dbgutil.FormatDisplay("Binaryfy", c.Binarify(img))
}

// OK
func TestCrop(t *testing.T) {
	dbgutil.FormatDisplay("Crop", c.Crop(img))
}

// OK
func TestLoadModule(t *testing.T) {
	train, err := c.LoadTrainModule(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	if err != nil {
		t.Fatal(err)
	}
	dbgutil.FormatDisplay("train", train)

	std, err := c.LoadStdModule(`d:\CCHelper\Golang\bin\Cleaned.dat`)
	if err != nil {
		t.Fatal(err)
	}
	dbgutil.FormatDisplay("std", std)
}

// OK
func TestAutoGenStdModule(t *testing.T) {
	//trainModule, err := c.LoadTrainModule(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//for _, zero := range trainModule['0'] {
	//	fmt.Println(zero)
	//}
	//t.Log(c.autoFindMatchest(trainModule['0']))
	stdModule, err := c.AutoGenStdModuleFromFile(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	if err != nil {
		t.Fatal(err)
	}
	for alpha, binimg := range stdModule {
		t.Logf("==[ %s ]==\n%s", string(alpha), binimg.String())
	}
}

/*
func TestManualGenStdModule(t *testing.T) {
	stdModule, err := c.ManualGenStdModule(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	if err != nil {
		t.Fatal(err)
	}
	for alpha, binimg := range stdModule {
		t.Logf("==[ %s ]==\n%s", string(alpha), binimg.String())
	}
}
无法使用测试框架测试此函数,单独测试
package main

import (
	"crmhelper_private/crm2"
	"fmt"
	"log"
)

const Threshole = 34000
const N = 6
var c *crm2.Captcha

func init() {
	img, err := crm2.ReadImg(`d:\CCHelper\Golang\bin\验证码\0.jpg`)
	if err != nil {
		log.Fatal(err)
	}
	c = crm2.NewCaptcha(img, Threshole, N)
}

func main() {
	stdModule, err := c.ManualGenStdModule(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	if err != nil {
		log.Fatal(err)
	}
	for alpha, binimg := range stdModule {
		fmt.Printf("==[ %s ]==\n%s", string(alpha), binimg.String())
	}
}
*/

func TestSaveTrainModule(t *testing.T) {
	train, err := c.LoadTrainModule(`d:\CCHelper\Golang\bin\Alphabet.dat`)
	if err != nil {
		t.Fatal(err)
	}
	err = c.SaveTrainModule(train, `d:\CCHelper\Golang\bin\Alphabet_Test.dat`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSaveStdModule(t *testing.T) {
	train, err := c.LoadStdModule(`d:\CCHelper\Golang\bin\Cleaned.dat`)
	if err != nil {
		t.Fatal(err)
	}
	err = c.SaveStdModule(train, `d:\CCHelper\Golang\bin\Cleaned_Test.dat`)
	if err != nil {
		t.Fatal(err)
	}
}

/*
func TestTrain(t *testing.T) {

}
无法测试,新文件测试
func main() {
	imgs := make([]image.Image, 5, 5)
	imgs[0], _ = crm2.ReadImg(`d:\CCHelper\Golang\bin\@jjjjj.jpg`)
	imgs[1], _ = crm2.ReadImg(`d:\CCHelper\Golang\bin\验证码\32.jpg`)
	imgs[2], _ = crm2.ReadImg(`d:\CCHelper\Golang\bin\验证码\12.jpg`)
	imgs[3], _ = crm2.ReadImg(`d:\CCHelper\Golang\bin\验证码\7.jpg`)
	imgs[4], _ = crm2.ReadImg(`d:\CCHelper\Golang\bin\验证码\55.jpg`)
	tm, err := c.Train(imgs, nil)
	if err != nil {
		log.Fatal(err)
	}
	dbgutil.FormatDisplay("", tm)
}
*/

func TestRecognize(t *testing.T) {
	_, err := c.LoadStdModule(`d:\CCHelper\Golang\bin\Cleaned.dat`)
	if err != nil {
		log.Fatal(err)
	}

	img, err := ReadImg(`d:\CCHelper\Golang\bin\@IIIIIIIIIII.jpg`)
	if err != nil {
		log.Fatal(err)
	}
	recognized := c.Recognize(img)
	t.Log(recognized)
}

func TestStdModuleCheck(t *testing.T) {
	std, err := c.LoadStdModule(`d:\CCHelper\Golang\bin\Cleaned.dat`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(c.StdModuleCheck(std, false))

	fmt.Println(c.StdModuleCheck(make(map[Alpha]BinaryImage), true))
}
