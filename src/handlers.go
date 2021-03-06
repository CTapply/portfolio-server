package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ResponseData struct {
	Message string  `json:"message"`
	Videos  []Video `json:"videos"`
}

type Video struct {
	VideoLocation     string `json:"videoLocation"`
	ThumbnailLocation string `json:"thumbnailLocation"`
}

func uploadHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Parse our multipart form, 10 << 32 specifies a maximum upload of 5 Gb
	request.ParseMultipartForm(10 << 32)
	file, handler, err := request.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	contentType := handler.Header.Get("Content-Type")

	// fmt.Println(formatRequest(request))

	// Write the file to disk
	fileEnding := getFileExtension(contentType)
	fileName := "upload-*" + fileEnding
	saveFile(fileName, file, request.FormValue("start"), request.FormValue("end"))

	data := ResponseData{Message: "Successfully Uploaded File"}
	writer.WriteHeader(http.StatusOK)
	encodingError := json.NewEncoder(writer).Encode(data)
	if encodingError != nil {
		fmt.Println(encodingError)
	}
}

func getVideosHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	rootDirectory := "."
	fileDirectory := "/resources/uploaded/"
	fileThumbnailDirectory := "/resources/thumbnails/"
	files := getVideos(rootDirectory + fileDirectory)

	data := ResponseData{Message: "Successfully Loaded Files", Videos: generateVideoResponseObject(files, request.Host, fileDirectory, fileThumbnailDirectory)}
	responseJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(responseJSON)
}

/**
 * Handler method for generating the json response object of the videoFiles that were loaded from disk
 */
func generateVideoResponseObject(videoFiles []os.FileInfo, host string, directoryPath string, thumbnailDirectory string) (response []Video) {
	// Set the security scheme for the video urls
	scheme := "https://"
	if env == "LOCAL" {
		scheme = "http://"
	}

	for _, file := range videoFiles {
		videoNameNoType := strings.TrimSuffix(file.Name(), ".mp4")
		response = append(response, Video{
			VideoLocation:     scheme + host + directoryPath + videoNameNoType + ".mp4",
			ThumbnailLocation: scheme + host + thumbnailDirectory + videoNameNoType + ".jpg",
		})
	}

	return response
}

/**
 * Handler method for just serving the index.html file
 */
func indexHandler(entrypoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entrypoint)
	}

	return http.HandlerFunc(fn)
}

/**
 * Handler method for responding to the health and readiness checks with a 204 code
 */
func healthCheckHandler(entrypoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	return http.HandlerFunc(fn)
}

// func getResourceHandler(writer http.ResponseWriter, request *http.Request) {
// 	fmt.Println("Sending File...")
// 	fmt.Println(request.URL.Path)
// 	fmt.Println(formatRequest(request))
// 	http.ServeFile(writer, request, "."+request.URL.Path)
// }
