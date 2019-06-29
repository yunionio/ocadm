package certs

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"k8s.io/klog"

	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
)

// NewCACertAndKey will generate a self signed CA.
func NewCACertAndKey(certSpec *certutil.Config) (*x509.Certificate, *rsa.PrivateKey, error) {

	caCert, caKey, err := pkiutil.NewCertificateAuthority(certSpec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failure while generating CA certificate and key")
	}

	return caCert, caKey, nil
}

// CreateCACertAndKeyFiles generates and writes out a given certificate authority.
// The certSpec should be one of the variables from this package.
func CreateCACertAndKeyFiles(certSpec *OnecloudCert, cfg *apis.InitConfiguration) error {
	if certSpec.CAName != "" {
		return errors.Errorf("this function should only be used for CAs, but cert %s has CA %s", certSpec.Name, certSpec.CAName)
	}
	klog.V(1).Infof("creating a new certificate authority for %s", certSpec.Name)

	certConfig, err := certSpec.GetConfig(cfg)
	if err != nil {
		return err
	}

	caCert, caKey, err := NewCACertAndKey(certConfig)
	if err != nil {
		return err
	}

	return writeCertificateAuthorithyFilesIfNotExist(
		cfg.OnecloudCertificatesDir,
		certSpec.BaseName,
		caCert,
		caKey,
	)
}

// NewCSR will generate a new CSR and accompanying key
func NewCSR(certSpec *OnecloudCert, cfg *apis.InitConfiguration) (*x509.CertificateRequest, *rsa.PrivateKey, error) {
	certConfig, err := certSpec.GetConfig(cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to retrieve cert configuration")
	}

	return pkiutil.NewCSRAndKey(certConfig)
}

// CreateCSR creates a certificate signing request
func CreateCSR(certSpec *OnecloudCert, cfg *apis.InitConfiguration, path string) error {
	csr, key, err := NewCSR(certSpec, cfg)
	if err != nil {
		return err
	}
	return writeCSRFilesIfNotExist(path, certSpec.BaseName, csr, key)
}

// CreateCertAndKeyFilesWithCA loads the given certificate authority from disk, then generates and writes out the given certificate and key.
// The certSpec and caCertSpec should both be one of the variables from this package.
func CreateCertAndKeyFilesWithCA(certSpec *OnecloudCert, caCertSpec *OnecloudCert, cfg *apis.InitConfiguration) error {
	if certSpec.CAName != caCertSpec.Name {
		return errors.Errorf("expected CAname for %s to be %q, but was %s", certSpec.Name, certSpec.CAName, caCertSpec.Name)
	}

	caCert, caKey, err := LoadCertificateAuthority(cfg.OnecloudCertificatesDir, caCertSpec.BaseName)
	if err != nil {
		return errors.Wrapf(err, "couldn't load CA certificate %s", caCertSpec.Name)
	}

	return certSpec.CreateFromCA(cfg, caCert, caKey)
}

// LoadCertificateAuthority tries to load a CA in the given directory with the given name.
func LoadCertificateAuthority(pkiDir string, baseName string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Checks if certificate authority exists in the PKI directory
	if !pkiutil.CertOrKeyExist(pkiDir, baseName) {
		return nil, nil, errors.Errorf("couldn't load %s certificate authority from %s", baseName, pkiDir)
	}

	// Try to load certificate authority .crt and .key from the PKI directory
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiDir, baseName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failure loading %s certificate authority", baseName)
	}

	// Make sure the loaded CA cert actually is a CA
	if !caCert.IsCA {
		return nil, nil, errors.Errorf("%s certificate is not a certificate authority", baseName)
	}

	return caCert, caKey, nil
}

// writeCertificateAuthorithyFilesIfNotExist write a new certificate Authority to the given path.
// If there already is a certificate file at the given path; ocadm tries to load it and check if the values in the
// existing and the expected certificate equals. If they do; ocadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func writeCertificateAuthorithyFilesIfNotExist(pkiDir string, baseName string, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {

	// If cert or key exists, we should try to load them
	if pkiutil.CertOrKeyExist(pkiDir, baseName) {

		// Try to load .crt and .key from the PKI directory
		caCert, _, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s certificate", baseName)
		}

		// Check if the existing cert is a CA
		if !caCert.IsCA {
			return errors.Errorf("certificate %s is not a CA", baseName)
		}

		// kubeadm doesn't validate the existing certificate Authority more than this;
		// Basically, if we find a certificate file with the same path; and it is a CA
		// kubeadm thinks those files are equal and doesn't bother writing a new file
		fmt.Printf("[certs] Using the existing %q certificate and key\n", baseName)
	} else {
		// Write .crt and .key files to disk
		fmt.Printf("[certs] Generating %q certificate and key\n", baseName)

		if err := pkiutil.WriteCertAndKey(pkiDir, baseName, caCert, caKey); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
	}
	return nil
}

// writeCertificateFilesIfNotExist write a new certificate to the given path.
// If there already is a certificate file at the given path; ocadm tries to load it and check if the values in the
// existing and the expected certificate equals. If they do; ocadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func writeCertificateFilesIfNotExist(pkiDir string, baseName string, signingCert *x509.Certificate, cert *x509.Certificate, key *rsa.PrivateKey, cfg *certutil.Config) error {

	// Checks if the signed certificate exists in the PKI directory
	if pkiutil.CertOrKeyExist(pkiDir, baseName) {
		// Try to load signed certificate .crt and .key from the PKI directory
		signedCert, _, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s certificate", baseName)
		}

		// Check if the existing cert is signed by the given CA
		if err := signedCert.CheckSignatureFrom(signingCert); err != nil {
			return errors.Errorf("certificate %s is not signed by corresponding CA", baseName)
		}

		// Check if the certificate has the correct attributes
		if err := validateCertificateWithConfig(signedCert, baseName, cfg); err != nil {
			return err
		}

		fmt.Printf("[certs] Using the existing %q certificate and key\n", baseName)
	} else {
		// Write .crt and .key files to disk
		fmt.Printf("[certs] Generating %q certificate and key\n", baseName)

		if err := pkiutil.WriteCertAndKey(pkiDir, baseName, cert, key); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
		if pkiutil.HasServerAuth(cert) {
			fmt.Printf("[certs] %s serving cert is signed for DNS names %v and IPs %v\n", baseName, cert.DNSNames, cert.IPAddresses)
		}
	}

	return nil
}

// writeKeyFilesIfNotExist write a new key to the given path.
// If there already is a key file at the given path; ocadm tries to load it and check if the values in the
// existing and the expected key equals. If they do; ocadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func writeKeyFilesIfNotExist(pkiDir string, baseName string, key *rsa.PrivateKey) error {

	// Checks if the key exists in the PKI directory
	if pkiutil.CertOrKeyExist(pkiDir, baseName) {

		// Try to load .key from the PKI directory
		_, err := pkiutil.TryLoadKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "%s key existed but it could not be loaded properly", baseName)
		}

		// kubeadm doesn't validate the existing certificate key more than this;
		// Basically, if we find a key file with the same path kubeadm thinks those files
		// are equal and doesn't bother writing a new file
		fmt.Printf("[certs] Using the existing %q key\n", baseName)
	} else {

		// Write .key and .pub files to disk
		fmt.Printf("[certs] Generating %q key and public key\n", baseName)

		if err := pkiutil.WriteKey(pkiDir, baseName, key); err != nil {
			return errors.Wrapf(err, "failure while saving %s key", baseName)
		}

		if err := pkiutil.WritePublicKey(pkiDir, baseName, &key.PublicKey); err != nil {
			return errors.Wrapf(err, "failure while saving %s public key", baseName)
		}
	}

	return nil
}

// writeCertificateAuthorithyFilesIfNotExist write a new CSR to the given path.
// If there already is a CSR file at the given path; ocadm tries to load it and check if it's a valid certificate.
// otherwise this function returns an error.
func writeCSRFilesIfNotExist(csrDir string, baseName string, csr *x509.CertificateRequest, key *rsa.PrivateKey) error {
	if pkiutil.CSROrKeyExist(csrDir, baseName) {
		_, _, err := pkiutil.TryLoadCSRAndKeyFromDisk(csrDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "%s CSR existed but it could not be loaded properly", baseName)
		}

		fmt.Printf("[certs] Using the existing %q CSR\n", baseName)
	} else {
		// Write .key and .csr files to disk
		fmt.Printf("[certs] Generating %q key and CSR\n", baseName)

		if err := pkiutil.WriteKey(csrDir, baseName, key); err != nil {
			return errors.Wrapf(err, "failure while saving %s key", baseName)
		}

		if err := pkiutil.WriteCSR(csrDir, baseName, csr); err != nil {
			return errors.Wrapf(err, "failure while saving %s CSR", baseName)
		}
	}

	return nil
}

type certKeyLocation struct {
	pkiDir     string
	caBaseName string
	baseName   string
	uxName     string
}

// validateCACert tries to load a x509 certificate from pkiDir and validates that it is a CA
func validateCACert(l certKeyLocation) error {
	// Check CA Cert
	caCert, err := pkiutil.TryLoadCertFromDisk(l.pkiDir, l.caBaseName)
	if err != nil {
		return errors.Wrapf(err, "failure loading certificate for %s", l.uxName)
	}

	// Check if cert is a CA
	if !caCert.IsCA {
		return errors.Errorf("certificate %s is not a CA", l.uxName)
	}
	return nil
}

// validateCACertAndKey tries to load a x509 certificate and private key from pkiDir,
// and validates that the cert is a CA
func validateCACertAndKey(l certKeyLocation) error {
	if err := validateCACert(l); err != nil {
		return err
	}

	_, err := pkiutil.TryLoadKeyFromDisk(l.pkiDir, l.caBaseName)
	if err != nil {
		return errors.Wrapf(err, "failure loading key for %s", l.uxName)
	}
	return nil
}

// validateSignedCert tries to load a x509 certificate and private key from pkiDir and validates
// that the cert is signed by a given CA
func validateSignedCert(l certKeyLocation) error {
	// Try to load CA
	caCert, err := pkiutil.TryLoadCertFromDisk(l.pkiDir, l.caBaseName)
	if err != nil {
		return errors.Wrapf(err, "failure loading certificate authority for %s", l.uxName)
	}

	return validateSignedCertWithCA(l, caCert)
}

// validateSignedCertWithCA tries to load a certificate and validate it with the given caCert
func validateSignedCertWithCA(l certKeyLocation, caCert *x509.Certificate) error {
	// Try to load key and signed certificate
	signedCert, _, err := pkiutil.TryLoadCertAndKeyFromDisk(l.pkiDir, l.baseName)
	if err != nil {
		return errors.Wrapf(err, "failure loading certificate for %s", l.uxName)
	}

	// Check if the cert is signed by the CA
	if err := signedCert.CheckSignatureFrom(caCert); err != nil {
		return errors.Wrapf(err, "certificate %s is not signed by corresponding CA", l.uxName)
	}
	return nil
}

// validatePrivatePublicKey tries to load a private key from pkiDir
func validatePrivatePublicKey(l certKeyLocation) error {
	// Try to load key
	_, _, err := pkiutil.TryLoadPrivatePublicKeyFromDisk(l.pkiDir, l.baseName)
	if err != nil {
		return errors.Wrapf(err, "failure loading key for %s", l.uxName)
	}
	return nil
}

// validateCertificateWithConfig makes sure that a given certificate is valid at
// least for the SANs defined in the configuration.
func validateCertificateWithConfig(cert *x509.Certificate, baseName string, cfg *certutil.Config) error {
	for _, dnsName := range cfg.AltNames.DNSNames {
		if err := cert.VerifyHostname(dnsName); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}
	for _, ipAddress := range cfg.AltNames.IPs {
		if err := cert.VerifyHostname(ipAddress.String()); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}
	return nil
}
