package general

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sme-api/app/env"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var supportedTypes = map[string]bool{
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"application/pdf": true,
}

// Read and save the excel file :
func ReadFile(file *multipart.FileHeader, path string) error {
	// get the temperary file path
	path = fmt.Sprintf("%s/%s", env.Config.Storage.Path, path)
	// check file size
	if file.Size > 8388608 {
		return errors.New("file size exceeds limit")
	}

	// read file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// check the directory exists
	CreateFolder(path)

	// to read the file type
	buff := make([]byte, 512)
	_, err = src.Read(buff)
	if err != nil {
		return err
	}

	src.Seek(0, 0)
	fileType := http.DetectContentType(buff)

	// check supported types
	if _, ok := supportedTypes[fileType]; !ok {
		return errors.New("unsupported file type")
	}

	// create a file under the directory
	dst, err := os.Create(path + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// copy the content to the destination
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// get the url
	// url, err := filepath.Abs(path + filename)
	// if err != nil {
	// 	return "Failed to get file path", err
	// }

	return nil
}

// Create folder
func CreateFolder(path string) {
	// check the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}

// RemoveFile :
func RemoveFile(path string) error {
	return os.Remove(path)
}

// FileExists :
func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Validate expire
func IsExpired(timeString string) bool {
	t := time.Now()
	rfcLayout := "2006-01-02T15:04:05Z07:00"
	timeNow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Format(time.RFC3339)
	now, _ := time.Parse(rfcLayout, timeNow)
	expTime, _ := time.Parse(rfcLayout, timeString)
	return expTime.Unix() > now.Unix()
}

// Get objectID from hex
func GetObjectID(id string) (*primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return &oid, nil
}
