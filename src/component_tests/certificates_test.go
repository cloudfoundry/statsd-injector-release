package component_tests_test

import (
	"log"
	"os"

	"code.cloudfoundry.org/tlsconfig/certtest"
)

func GenerateCertKey(name string, ca *certtest.Authority) (string, string) {
	cert, err := ca.BuildSignedCertificate(name, certtest.WithDomains(name))
	if err != nil {
		log.Fatal(err)
	}
	certBytes, keyBytes, err := cert.CertificatePEMAndPrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	return createTempFile(certBytes), createTempFile(keyBytes)
}

func GenerateCA() (*certtest.Authority, string) {
	ca, err := certtest.BuildCA("metron")
	if err != nil {
		log.Fatal(err)
	}

	caBytes, err := ca.CertificatePEM()
	if err != nil {
		log.Fatal(err)
	}

	return ca, createTempFile(caBytes)
}

func createTempFile(contents []byte) string {
	tmpfile, err := os.CreateTemp("", "")

	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(contents)); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()
}
