package optimize

import (
	"errors"
	"log"
	"net/http"
	"os"
)

// ErrFormat indicates that optimize encountered an unknown format.
var ErrFormat = errors.New("optimize: unknown format")

type OptimizeSummary struct {
	// Size in bytes before passing in the optimization function
	SizeBefore int64
	// Size in bytes after passing in the optimization function
	SizeAfter int64
	// Size in bytes after passing in the optimization function
	SizeLossy int64
}

// Optimizer is interface for optimization
type optimizer struct {
	name, mimeType string

	// Uses a optimize function on the source and saves to the destination
	// Returns a summary of the optimization
	optimizeFile func(string, string, string) (OptimizeSummary, error)

	// Uses a optimization function on input data
	// Returns optimized data along with a summary of optimization
	optimizeData func([]byte, string) ([]byte, OptimizeSummary, error)
}

var optimizers []optimizer

// RegisterOptimizer registers an image format for use by Decode.
// Name is the name of the format, like "jpeg" or "png".
// MimeType is the mime type that identifies the format's encoding.
// OptimizeFile is the function that optimizes from a file to another file, returning the optimization summary.
// OptimizeData is the function that optimizes a byte array source and returns a byte array source.
func RegisterOptimizer(name, mimetype string, optimizeFile func(string, string, string) (OptimizeSummary, error), optimizeData func([]byte, string) ([]byte, OptimizeSummary, error)) {
	log.Println("Registered optimizer: " + name)
	optimizers = append(optimizers, optimizer{name, mimetype, optimizeFile, optimizeData})
}

func match(mimeType string) optimizer {
	for _, c := range optimizers {
		if c.mimeType == mimeType {
			return c
		}
	}
	return optimizer{}
}

func detectFileMimeType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	// Create a buffer to read the first 512 bytes, which is enough to detect the mimetype
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

// File tries to optimize a file and save it to another file
func File(src string, dst string, lossy string) (OptimizeSummary, error) {
	mimeType, err := detectFileMimeType(src)
	if err != nil {
		return OptimizeSummary{}, err
	}
	c := match(mimeType)
	log.Println("Match:")
	log.Println(c)
	if c.optimizeFile == nil {
		return OptimizeSummary{}, ErrFormat
	}
	return c.optimizeFile(src, dst, lossy)
}

// Data tries to optimize the data and return it
func Data(src []byte, lossy string) ([]byte, OptimizeSummary, error) {
	mimeType := http.DetectContentType(src)
	c := match(mimeType)
	if c.optimizeData == nil {
		return nil, OptimizeSummary{}, ErrFormat
	}
	return c.optimizeData(src, lossy)
}

func fileStat(src string) (int64, error) {
	stat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}
