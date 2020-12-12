// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Generate a self-signed X.509 certificate for a TLS server. Outputs to
// 'certificates.pem' and 'key.pem' and will overwrite existing files.

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

// certificates duration should be 10 year, we add a 10 year grace
const validDuration = 20*365*24*time.Hour
const uncopiedOrg = "uncopied"
const uncopiedCountry = "France"
const uncopiedLocality = "Versailles"
const uncopiedPostalCode = "78000"
const uncopiedStreetAddress = "BP XXX"
const uncopiedDomainOrg = "uncopied.org"
const uncopiedDomainArt = "uncopied.art"
const emailAddress = "contact@uncopied.org"
const emailAddressPresident = "elian.carsenat@uncopied.art"
const rsaBits = 2048

// see also http://www.inanzzz.com/index.php/post/1nw3/data-encryption-and-decryption-with-x509-public-and-private-key-example-in-golang
func main() {


	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template

	notBefore := time.Now()
	notAfter := notBefore.Add(validDuration)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	keyUsage |= x509.KeyUsageKeyEncipherment
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{uncopiedOrg},
			Country: []string{uncopiedCountry},
			Locality: []string{uncopiedLocality},
			PostalCode: []string{uncopiedPostalCode},
			StreetAddress: []string{uncopiedStreetAddress},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

	}
	template.DNSNames = append(template.DNSNames, uncopiedDomainOrg)
	template.DNSNames = append(template.DNSNames, uncopiedDomainArt)
	template.EmailAddresses = append(template.EmailAddresses,emailAddress)
	template.EmailAddresses = append(template.EmailAddresses,emailAddressPresident)
	//whether this certificates should be its own Certificate Authority
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create("certificates.pem")
	if err != nil {
		log.Fatalf("Failed to open certificates.pem for writing: %v", err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to certificates.pem: %v", err)
	}

	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing certificates.pem: %v", err)
	}
	log.Print("wrote certificates.pem\n")

	keyOut, err := os.OpenFile("keyPrivate.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open key.pem for writing: %v", err)
		return
	}

	keyPubOut, err := os.OpenFile("keyPublic.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open key.pem for writing: %v", err)
		return
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing key.pem: %v", err)
	}

	pubBytes := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	if err != nil {
		log.Fatalf("Unable to marshal public key: %v", err)
	}

	if err := pem.Encode(keyPubOut, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %v", err)
	}

	if err := keyPubOut.Close(); err != nil {
		log.Fatalf("Error closing key.pem: %v", err)
	}

	log.Print("wrote key.pem\n")
	someUuid := uuid.New().String()
	secretOut, err := os.Create("secret.pem")
	if err != nil {
		log.Fatalf("Failed to open secret.pem for writing: %v", err)
	}
	_,err := secretOut.Write(someUuid)
	if err != nil {
		log.Fatalf("Failed to write data to secret.pem: %v", err)
	}
	if err := secretOut.Close(); err != nil {
		log.Fatalf("Error closing secret.pem: %v", err)
	}

}