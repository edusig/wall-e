package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nickalie/go-mozjpegbin"

	"github.com/justinas/alice"
	"github.com/satori/go.uuid"
)

type FileDetails struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type UploadResponse struct {
	Source     *FileDetails `json:"source"`
	Compressed *FileDetails `json:"compressed"`
	FileType   string       `json:"fileType"`
}

func compressJPEG(source string, destination string) (int64, error) {
	var size int64
	in, err := os.Open(source)
	if err != nil {
		return size, err
	}
	log.Println(source)
	img, err := jpeg.Decode(in)
	if err != nil {
		return size, err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return size, err
	}
	defer out.Close()
	if err := mozjpegbin.Encode(out, img, nil); err != nil {
		log.Println("MOZJPEG ERROR")
		return size, err
	}
	stat, err := out.Stat()
	if err != nil {
		return size, err
	}
	size = stat.Size()
	return size, nil
}

func mimeTypeToExt(mime string) string {
	mimeExtension := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
	}
	return mimeExtension[mime]
}

func generateFileToken(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer f.Close()
	hasher := sha1.New()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func isAllowedMimeType(fileType string) bool {
	types := [2]string{"image/jpeg", "image/png"}
	allowed := false
	for i := 0; i < len(types); i++ {
		if fileType == types[i] {
			allowed = true
			break
		}
	}
	return allowed
}

func logginHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v\n", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func readFileStream(w http.ResponseWriter, r *http.Request, dest string) string {
	var fileType string
	mr, err := r.MultipartReader()
	if err != nil {
		log.Print("MultipartReader Error: ")
		log.Println(err)
		return ""
	}
	// length := r.ContentLength
	// Start Reading File
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		var read int64
		// var p float32
		dst, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Print("Open File Error: ")
			log.Println(err)
			return ""
		}
		defer dst.Close()

		for {
			buffer := make([]byte, 100000)
			cBytes, err := part.Read(buffer)
			if err == io.EOF {
				break
			}
			read = read + int64(cBytes)
			// p = float32(read) / float32(length) * 100
			// log.Printf("progress: %f\n", p)

			// After reading the first part of the file, checks the type
			if fileType == "" {
				fileType = http.DetectContentType(buffer[0:cBytes])
				if !isAllowedMimeType(fileType) {
					http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
					log.Printf("Trying to upload a file with mime type: %s\n", fileType)
					return ""
				}
			}
			dst.Write(buffer[0:cBytes])
		}
	}
	return fileType
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		tempFileUUID := uuid.NewV4().String()
		tempPath := "./temp/" + tempFileUUID

		mimeType := readFileStream(w, r, tempPath)
		if mimeType == "" {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fileType := mimeTypeToExt(mimeType)

		// Generates final folder based on random uuid
		finalFolder := uuid.NewV4().String()
		finalPath := "/uploads/" + finalFolder

		// Creates new destionation folder
		if err := os.Mkdir("./public/"+finalPath, 0777); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Generates final filename based on SHA-1 checksum
		token := generateFileToken(tempPath)
		sourceFilePath := finalPath + "/" + token + "." + fileType

		// Moves the original file from temp path to the final destination
		if err := os.Rename(tempPath, "./public/"+sourceFilePath); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		//Compress File
		compressedSize, err := compressJPEG("."+sourceFilePath, tempPath)
		if err != nil {
			log.Println("Compression Error: ")
			log.Print(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Generates final filename based on SHA-1 checksum
		token = generateFileToken(tempPath)
		compressFilePath := finalPath + "/" + token + "." + fileType

		// Moves the compressed file from temp path to the final destination
		if err := os.Rename(tempPath, "."+compressFilePath); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Returns file stats and summary of compression
		sourceFile := FileDetails{URL: sourceFilePath, Size: r.ContentLength}
		compressedFile := FileDetails{URL: compressFilePath, Size: compressedSize}
		fileCompressed := UploadResponse{Source: &sourceFile, Compressed: &compressedFile, FileType: mimeType}
		log.Println(fileCompressed)
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(fileCompressed)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {
	commonHandlers := alice.New(logginHandler, recoverHandler)
	http.Handle("/", commonHandlers.Then(http.FileServer(http.Dir("./public"))))
	http.Handle("/upload", commonHandlers.ThenFunc(uploadHandler))
	log.Println("Starting server at http://localhost:8000/")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}
