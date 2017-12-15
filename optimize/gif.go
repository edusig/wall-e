package optimize

import (
	"os/exec"
)

func GifFile(src string, dst string, lossy string) (OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	var err error
	opt.SizeBefore, err = fileStat(src)
	if err != nil {
		return opt, err
	}

	cmd := exec.Command("gifsicle", "-O3", "-o", dst, src)
	if err := cmd.Run(); err != nil {
		return opt, err
	}

	opt.SizeAfter, err = fileStat(dst)
	if err != nil {
		return opt, err
	}

	return opt, nil
}

func GifData(src []byte, lossy string) ([]byte, OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	return nil, opt, nil
}

func init() {
	RegisterOptimizer("gif", "image/gif", GifFile, nil)
}
