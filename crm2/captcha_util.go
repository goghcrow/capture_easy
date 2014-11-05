// todo 加入压缩与解压,加入读取与写入文件两个包装函数
package crm2

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/jpeg"
	"os"
)

// !panic 计算字节数组汉明距离  n <= 5 true n > 10 false
func Hamming(a, b []byte) (n int) {
	l := len(a)
	if l != len(b) {
		panic(fmt.Errorf("Hamming len(a)[%d] != len(b)[%d]", l, len(b)))
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			n += 1
		}
	}
	return
}

// 从文件读取imgage
func ReadImg(imgFile string) (image.Image, error) {
	f, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}
	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// @from http://golanghome.com/post/346
func ByteEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// @from http://golanghome.com/post/346
// to 必须为point
func ByteDecode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}
