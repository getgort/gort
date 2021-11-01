/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func generateKeyBytes() ([]byte, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, Inc."},
			StreetAddress: []string{"2020 5th Avenue"},
			Locality:      []string{"New York"},
			Province:      []string{"NY"},
			PostalCode:    []string{"10000"},
			Country:       []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return caPEM.Bytes(), caPrivKeyPEM.Bytes(), nil
}

func GenerateTemporaryTLSKeys() (string, string, error) {
	filemode := os.FileMode(0600)

	certBytes, keyBytes, err := generateKeyBytes()
	if err != nil {
		return "", "", err
	}

	certFile, err := os.CreateTemp("", "gort.c.")
	if err != nil {
		return "", "", err
	}
	if err := certFile.Chmod(filemode); err != nil {
		return "", "", err
	}
	if _, key := certFile.Write(certBytes); key != nil {
		return "", "", err
	}

	keyFile, err := os.CreateTemp("", "gort.k.")
	if err != nil {
		return "", "", err
	}
	if err := keyFile.Chmod(filemode); err != nil {
		return "", "", err
	}
	if _, key := keyFile.Write(keyBytes); key != nil {
		return "", "", err
	}

	return certFile.Name(), keyFile.Name(), nil
}
