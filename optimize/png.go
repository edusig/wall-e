package optimize

import (
	"os/exec"
)

func PngFile(src string, dst string, lossy string) (OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	var err error
	opt.SizeBefore, err = fileStat(src)
	if err != nil {
		return opt, err
	}

	cmd := exec.Command("optipng", "-out="+dst, src)
	if err := cmd.Run(); err != nil {
		return opt, err
	}

	if lossy != "" {
		cmd = exec.Command("pngquant", "--output="+lossy, src)
		if err := cmd.Run(); err != nil {
			return opt, err
		}
		opt.SizeLossy, err = fileStat(lossy)
	}

	opt.SizeAfter, err = fileStat(dst)
	if err != nil {
		return opt, err
	}

	return opt, nil
}

func PngData(src []byte, lossy string) ([]byte, OptimizeSummary, error) {
	var opt = OptimizeSummary{}
	return nil, opt, nil
}

func init() {
	RegisterOptimizer("png", "image/png", PngFile, nil)
}
