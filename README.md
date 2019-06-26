# https-reverse-shell-golang

A HTTPS reverse shell implemented in the Go programming language. 

## Instructions

In order to build, you need to create a server cert/key pair. Then put the ```server.cert``` and  ```server.key```  into ```c2/resources``` and the ```server.cert``` as well into ```payload/resources```.

You can create the pair for example like this:

```$ openssl genrsa -out server.key 2048```

```$ openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650```

Now run ```$ go install payload.go``` and ```$ go install c2.payload``` and the binaries will be created in your ```$GOBIN``` folder.

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


