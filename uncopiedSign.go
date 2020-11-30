package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// see also http://www.inanzzz.com/index.php/post/1nw3/data-encryption-and-decryption-with-x509-public-and-private-key-example-in-golang
func main() {
	var privPEMData = []byte(`
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCjwdu1Mh8d3I08
oBuEOJVePgXj1wKyzu3qxjyVmCgikSB9XRegy/DPpX/n4uRqxVB4iPZXflGu7sch
FhCMdfxb4byBI9JF7p4rK3xHlab9a4EkUdvOEr28zTNgtwkmme6CDFbhtxXFQSjd
fPQ+BKYA0t2x8IHqAoI9dBCHLTj9BN7HPIagNFLv8gMM9SXEzA1FGHxp7OhoOuHZ
ltfAOKI1Suwvd51/oHSUojRcX0LoTeLJovvX/lJl1Mu72eW1RBKdBv7Sxk4sYmkb
1UCitMnCo+ZCqdb/8qGbjg+S7qelRb6jNj/H+brcHIUw3uMbz2goYP/NbCGU/4uG
XWYvXX11AgMBAAECggEBAISPnYdkd4P40exNv3idRWzw0FvL5cdRc48lwk1myraQ
vLg+762e6eVtl8jjBvzXlXi9hoz1GLJ/YHsMHYFW0V6fsbTohoNN0oQnw4c/QdrL
d9Mq4MBEs4tuoTSddq7k1Qo5aut1Bg6T3LzPNfguUyM/j29HviLsvPl6RxbmKMfI
KfBmGs5Enxvyp3DJkP1CvQDmEuB2kW8fGXYbVy+x3plACOO/OSBvbrTWp4fGrCo4
mNWMpAK/MU208MlCUNMH4PQXJgV7uXm0aRiTDxDWyYtSH056muIqFwfkrGzZJMkt
93MzhnU36bXOX1mQsAPYMmjxpJfaScXeL2wX7tHb8oECgYEA0OLqz1Sz9kA9vZ/y
x8Bo/PejKibJzY9JQ0n2Jqs9Fy6dcIA7e5KiwU4SEbXzWISHFsXV9uRZSWR+Ci65
XEOApFxgrAh/oYLru7zR1pYrjr4ez+cc6m685rv7ryocsssCq9LoxxECsCA5Io+n
cant0jbU83P+H9VWvauPak+izEkCgYEAyLEwQB+rdNSGBujw/K8JICavdaFXoqCo
OyafN/LLQha5XNxN2O2MTmctMUTl7DMmzwjtkvBZEB+OAVBZmjEeQ92rY7vtKFxE
OmKuEDgDnKLVwZm2YLjLD/co98RXf3KeBUcIdaRinbpwsDQ9A4S3/6xjbgdTPyWh
j0lZZ/mKr80CgYBjrgVzTu5Z8qoD1VIbtFvla573PG9MorXJYIAQT+LlLx9+UhMQ
kxcLu9+vh+5KLWPxoBLMsIdTGJt07HsT5jp7NIIFVkDhqAIqIp7YEe1TPrKhb55C
2PlX+hjOq//p6iqqKAlhBWMM/TOGpJq5COguSnAwhQed1UaBWF8l0j7T0QKBgGVc
eI4qcKJVJEwhInW8wdMnNr8meeh9U/psC0ZqrhX2/C/WZMsHTzHaEo0ryyR8wUEX
tUXddl4aUdKADoE+BZcpQgLhS2pzD1KdvGQcplZaN7PMOrynGIg7wMlCtR59eSoZ
MkCYgeY/3+Jev+IjCftrydwsfvMJwotn9Gv7MPyRAoGAdhk++VzYWrz65FL36Btj
NbrEVeXVA3z+ByoUXR+CKIInDSMEf+FDp8kE5EIhVd3UT0pEgLFjB94ZTbytoShD
qx+LF1b6Ndg156XG1xEThZB58zdZJdEz96NqGFkXHEBUCI9E7/j2OWLUeJ6eOUwo
wGS3ha8R0NfDrjnPv0Vrfgw=
-----END PRIVATE KEY-----
`)


	block, rest := pem.Decode(privPEMData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got a %T, with remaining data: %q", privateKey, rest)

	secretMessage := []byte("send reinforcements, we're going to advance")
	label := []byte("orders")

	var privateKeyRsa = privateKey.(*rsa.PrivateKey)

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&privateKeyRsa.PublicKey,
		secretMessage,
		label)

	if err != nil {
		panic(err)
	}

	fmt.Println("encrypted bytes: ", encryptedBytes)
	fmt.Println(hex.EncodeToString(encryptedBytes))
	encrypted := hex.EncodeToString(encryptedBytes)
	fmt.Println(encrypted)

	// now decrypt it :
	ciphertext, _ := hex.DecodeString(encrypted)

	// crypto/rand.Reader is a good source of entropy for blinding the RSA
	// operation.
	rng := rand.Reader

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, privateKeyRsa, ciphertext, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
		return
	}

	fmt.Printf("Plaintext: %s\n", string(plaintext))

}
