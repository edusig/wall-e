package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/satori/go.uuid"

	"github.com/edusig/wall-e/handler"
	"github.com/edusig/wall-e/optimize"
)

type FileDetails struct {
	URL         string  `json:"url"`
	Size        int64   `json:"size"`
	SizeDiff    int64   `json:"sizeDiff,omitempty"`
	PercentDiff float64 `json:"percentDiff,omitempty"`
}

type Upload struct {
	Source     *FileDetails `json:"source"`
	Compressed *FileDetails `json:"compressed"`
	Lossy      *FileDetails `json:"lossy,omitempty"`
	FileType   string       `json:"fileType"`
}

var allowedMimeTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
}

var errorMimeTypeNotAllowed = errors.New("Mimetype not recognized")
var errorMimeTypeEmpty = errors.New("Mimetype empty")
var errorFileTypeNotAllowed = errors.New("File type not allowed")
var errorMethodNowAllowed = errors.New(http.StatusText(http.StatusMethodNotAllowed))
var errorNoCompression = errors.New("Could not reduce file size")

func mimeTypeToExt(mime string) (string, error) {
	if mime == "" {
		return mime, errorMimeTypeEmpty
	}
	if mimeType, ok := allowedMimeTypes[mime]; ok {
		return mimeType, nil
	}
	return "", errors.New("Mimetype not recognized")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func isAllowedMimeType(fileType string) bool {
	types := []string{"image/jpeg", "image/png", "image/gif"}
	return stringInSlice(fileType, types)
}

func copyFormFile(r *http.Request, dest string) (string, error) {
	var fileType string

	// Parse using 32mb
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return fileType, err
	}

	// Opens uploaded file
	file, _, err := r.FormFile("upload[file]")
	if err != nil {
		return fileType, err
	}
	defer file.Close()

	// Create a buffer to read the first 512 bytes, which is enough to detect the mimetype
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return fileType, err
	}
	file.Seek(0, 0)

	fileType = http.DetectContentType(buffer)
	if !isAllowedMimeType(fileType) {
		log.Printf("Trying to upload a file with mime type: %s\n", fileType)
		return fileType, errorFileTypeNotAllowed
	}

	// Creates the temp file to store the upload
	source, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return fileType, err
	}
	defer source.Close()

	// Copies the uploaded file to the temp file destination
	if _, err := io.Copy(source, file); err != nil {
		return fileType, err
	}

	return fileType, nil
}

func deleteIfExists(filePath string) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Could not remove file: %s", filePath)
			return
		}
	}
}

func uploadHandler(env *handler.Env, w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		var response handler.Response
		var lossyFile string
		var lossyFilePath string
		var lossyFileDetails FileDetails

		tempPath := "./temp/"
		sourceFile := uuid.NewV4().String()
		mimeType, err := copyFormFile(r, tempPath+sourceFile)
		if err != nil {
			if err == errorFileTypeNotAllowed {
				return handler.FlowError{Err: err, Detail: "Supported file formats are: JPEG, JPG and PNG", Code: 0}
			}
			return err
		}
		// If something failed delete the temporary file
		defer deleteIfExists(tempPath + sourceFile)

		generateLossy := mimeType == "image/png"
		fileType, err := mimeTypeToExt(mimeType)
		if err != nil {
			return err
		}

		compressedFile := uuid.NewV4().String()
		if generateLossy {
			lossyFile = uuid.NewV4().String()
		}

		//Compress File
		opt, err := optimize.File(tempPath+sourceFile, tempPath+compressedFile, tempPath+lossyFile)
		if err != nil {
			return err
		}

		// If something failed delete the temporary file
		defer deleteIfExists(tempPath + compressedFile)

		// Checks if uploaded file is already compressed by comparing file size with the compressed file
		isAlreadyCompressed := false
		// Minimal compression percentage to consider successfull
		minDiffPercentage := 1.0
		compressedDiff := opt.SizeBefore - opt.SizeAfter
		var compressedP float64
		var lossyDiff int64
		var lossyP float64
		if compressedDiff > 0 {
			compressedP = float64(100 - (opt.SizeAfter / opt.SizeBefore * 100))
			if compressedP < minDiffPercentage {
				isAlreadyCompressed = true
			}
		} else {
			isAlreadyCompressed = true
		}

		if generateLossy {
			lossyDiff := opt.SizeBefore - opt.SizeLossy
			if lossyDiff > 0 {
				lossyP = float64(100 - (opt.SizeLossy / opt.SizeBefore * 100))
				if lossyP >= minDiffPercentage {
					isAlreadyCompressed = false
				}
			}
		}

		if !isAlreadyCompressed {
			// Generates final folder based on random uuid
			finalPath := "/uploads/" + uuid.NewV4().String()

			sourceFilePath := finalPath + "/" + sourceFile + "." + fileType
			compressFilePath := finalPath + "/" + compressedFile + "." + fileType
			if generateLossy {
				lossyFilePath = finalPath + "/" + lossyFile + "." + fileType
			}

			// Creates new destination folder
			if err := os.Mkdir("./public/"+finalPath, 0777); err != nil {
				return err
			}
			// Moves files from temp path to the final destination
			if err := os.Rename(tempPath+sourceFile, "./public/"+sourceFilePath); err != nil {
				return err
			}
			if err := os.Rename(tempPath+compressedFile, "./public/"+compressFilePath); err != nil {
				return err
			}
			if generateLossy {
				if err := os.Rename(tempPath+lossyFile, "./public/"+lossyFilePath); err != nil {
					return err
				}
			}

			sourceFileDetails := FileDetails{URL: sourceFilePath, Size: opt.SizeBefore}
			compressedFileDetails := FileDetails{URL: compressFilePath, Size: opt.SizeAfter, SizeDiff: compressedDiff, PercentDiff: compressedP}
			uploadResponse := Upload{Source: &sourceFileDetails, Compressed: &compressedFileDetails, FileType: mimeType}
			if generateLossy {
				lossyFileDetails = FileDetails{URL: lossyFilePath, Size: opt.SizeLossy, SizeDiff: lossyDiff, PercentDiff: lossyP}
				uploadResponse.Lossy = &lossyFileDetails
			}
			response = handler.Response{Result: &uploadResponse, Success: true}
			// Returns file stats and summary of compression
			log.Println(response)
			w.Header().Set("Content-type", "application/json")
			json.NewEncoder(w).Encode(response)
			return nil
		}
		return handler.FlowError{Err: errorNoCompression, Detail: "The compression could not reduce file size. Either your file is already compressed", Code: 1, Type: "NO_COMPRESSION"}
	}
	return handler.FlowError{Code: http.StatusMethodNotAllowed, Err: errorMethodNowAllowed}
}

func main() {
	env := &handler.Env{}
	http.Handle("/", handler.LogginHandler(handler.RecoverHandler(http.FileServer(http.Dir("./public")))))
	http.Handle("/upload", handler.LogginHandler(handler.RecoverHandler(handler.Handler{env, uploadHandler})))
	log.Println("Starting server at http://localhost:8000/")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}
