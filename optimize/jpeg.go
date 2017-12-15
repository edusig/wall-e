package optimize

import (
	"bufio"
	"bytes"
	"log"

	mozjpegbin "github.com/nickalie/go-mozjpegbin"
)

func JpegFile(src string, dst string, lossy string) (OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	var err error
	opt.SizeBefore, err = fileStat(src)
	if err != nil {
		return opt, err
	}
	err = mozjpegbin.NewJpegTran().InputFile(src).Progressive(true).OutputFile(dst).Run()
	mozjpegbin.NewJpegTran().CopyNone()
	if err != nil {
		log.Println("MOZJPEG ERROR")
		return opt, err
	}
	opt.SizeAfter, err = fileStat(dst)
	if err != nil {
		return opt, err
	}
	return opt, nil
}

func JpegData(src []byte, lossy string) ([]byte, OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	opt.SizeBefore = int64(len(src))
	var res bytes.Buffer
	w := bufio.NewWriter(&res)
	err := mozjpegbin.NewJpegTran().Input(bytes.NewReader(src)).Progressive(true).Output(w).Run()
	mozjpegbin.NewJpegTran().CopyNone()
	if err != nil {
		return nil, opt, err
	}
	w.Flush()
	opt.SizeAfter = int64(res.Len())
	return nil, opt, nil
}

func init() {
	RegisterOptimizer("jpeg", "image/jpeg", JpegFile, JpegData)
}
