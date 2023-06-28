package aws

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"sme-api/app/env"
	"sme-api/app/kit/general"
	"sme-api/app/response"
	"sme-api/app/response/errcode"
)

// PushDocBucket :
func PushDocBucket(bucketPath string, file *multipart.FileHeader) (string, int, *response.Exception) {
	err := general.ReadFile(file, bucketPath)
	if err != nil {
		return "", http.StatusBadRequest, &response.Exception{
			Code:    errcode.InvalidFile,
			Error:   err,
			Message: "System failed to read the file",
		}
	}
	// get the temperary file path
	path := fmt.Sprintf("%s/%s", env.Config.Storage.Path, bucketPath)
	// read file from local storage
	savedFile, err := os.Open(path + file.Filename) // For read access.
	if err != nil {
		return "", http.StatusBadRequest, &response.Exception{
			Code:    errcode.InvalidFile,
			Error:   err,
			Message: "System failed to read the file",
		}
	}

	// push and create signed url
	url, err := PushDocS3Bucket(savedFile, bucketPath, file.Filename)
	if err != nil {
		return "", http.StatusBadRequest, &response.Exception{
			Code:    errcode.InvalidFile,
			Error:   err,
			Message: "System failed to push the file to bucket",
		}
	}
	return url, http.StatusOK, nil
}
