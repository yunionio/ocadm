package certs

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/pkiutil"
)

type configMutatorsFunc func(configuration *apis.InitConfiguration, config *certutil.Config) error

// OnecloudCert represents a certificate that Onecloud will create to function properly.
type OnecloudCert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	// Some attributes will depend on the InitConfiguration, only known at runtime.
	// These functions will be run in series, passed both the InitConfiguration and a cert Config.
	configMutators []configMutatorsFunc
	config         certutil.Config
}

func (k *OnecloudCert) GetConfig(ic *apis.InitConfiguration) (*certutil.Config, error) {
	for _, f := range k.configMutators {
		if err := f(ic, &k.config); err != nil {
			return nil, err
		}
	}

	return &k.config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (k *OnecloudCert) CreateFromCA(ic *apis.InitConfiguration, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", k.Name)
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return err
	}
	err = writeCertificateFilesIfNotExist(
		ic.OnecloudCertificatesDir,
		k.BaseName,
		caCert,
		cert,
		key,
		cfg,
	)

	if err != nil {
		return errors.Wrapf(err, "failed to write or validate certificate %q", k.Name)
	}

	return nil
}

func (k *OnecloudCert) CreateAsCA(ic *apis.InitConfiguration) (*x509.Certificate, *rsa.PrivateKey, error) {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't get configuration for %q CA certificate", k.Name)
	}
	caCert, caKey, err := NewCACertAndKey(cfg)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't generate %q CA certificate", k.Name)
	}

	err = writeCertificateAuthorithyFilesIfNotExist(
		ic.OnecloudCertificatesDir,
		k.BaseName,
		caCert,
		caKey,
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't write out %q CA certificate", k.Name)
	}

	return caCert, caKey, nil
}

// CertificateTree is represents a one-level-deep tree, mapping a CA to the certs that depend on it.
type CertificateTree map[*OnecloudCert]Certificates

// CreateTree creates the CAs, certs signed by the CAs, and rites them all to disk.
func (t CertificateTree) CreateTree(ic *apis.InitConfiguration) error {
	for ca, leaves := range t {
		cfg, err := ca.GetConfig(ic)
		if err != nil {
			return err
		}

		var caKey *rsa.PrivateKey

		caCert, err := pkiutil.TryLoadCertFromDisk(ic.OnecloudCertificatesDir, ca.BaseName)
		if err == nil {
			// Cert exists already, make sure it's valid
			if !caCert.IsCA {
				return errors.Errorf("certificate %q is not a CA", ca.Name)
			}
			// Try and load a CA Key
			caKey, err = pkiutil.TryLoadKeyFromDisk(ic.OnecloudCertificatesDir, ca.BaseName)
			if err != nil {
				// If there's no CA key, make sure every certificate exists.
				for _, leaf := range leaves {
					cl := certKeyLocation{
						pkiDir:   ic.OnecloudCertificatesDir,
						baseName: leaf.BaseName,
						uxName:   leaf.Name,
					}
					if err := validateSignedCertWithCA(cl, caCert); err != nil {
						return errors.Wrapf(err, "could not load expected certificate %q or validate the existence of key %q for it", leaf.Name, ca.Name)
					}
				}
				continue
			}
			// CA key exists; just use that to create new certificates.
		} else {
			// CACert doesn't already exist, create a new cert and key.
			caCert, caKey, err = NewCACertAndKey(cfg)
			if err != nil {
				return err
			}

			err = writeCertificateAuthorithyFilesIfNotExist(
				ic.OnecloudCertificatesDir,
				ca.BaseName,
				caCert,
				caKey,
			)
			if err != nil {
				return err
			}
		}

		for _, leaf := range leaves {
			if err := leaf.CreateFromCA(ic, caCert, caKey); err != nil {
				return err
			}
		}
	}
	return nil
}

// CertificateMap is a flat map of certificates, keyed by Name.
type CertificateMap map[string]*OnecloudCert

// CertTree returns a one-level-deep tree, mapping a CA cert to an array of certificates that should be signed by it.
func (m CertificateMap) CertTree() (CertificateTree, error) {
	caMap := make(CertificateTree)

	for _, cert := range m {
		if cert.CAName == "" {
			if _, ok := caMap[cert]; !ok {
				caMap[cert] = []*OnecloudCert{}
			}
		} else {
			ca, ok := m[cert.CAName]
			if !ok {
				return nil, errors.Errorf("certificate %q references unknown CA %q", cert.Name, cert.CAName)
			}
			caMap[ca] = append(caMap[ca], cert)
		}
	}

	return caMap, nil
}

// Certificates is a list of Certificates that Kubeadm should create.
type Certificates []*OnecloudCert

// AsMap returns the list of certificates as a map, keyed by name.
func (c Certificates) AsMap() CertificateMap {
	certMap := make(map[string]*OnecloudCert)
	for _, cert := range c {
		certMap[cert.Name] = cert
	}

	return certMap
}

// GetDefaultCertList returns all of the certificates ocadm requires to function.
func GetDefaultCertList() Certificates {
	return Certificates{
		&OcadmCertRootCA,
		&OcadmCertKeystoneServer,
		&OcadmClimcCertClient,
		&OcadmCertRegionServer,
	}
}

var (
	// OcadmCertRootCA is the definition of the onecloud Root CA for services
	OcadmCertRootCA = OnecloudCert{
		Name:     "ca",
		LongName: "self-signed onecloud CA to provision identities for other service components",
		BaseName: constants.CACertAndKeyBaseName,
		config: certutil.Config{
			CommonName: "onecloud",
		},
	}
	// OcadmCertKeystoneServer is the definition of the cert used to serve the identity service.
	OcadmCertKeystoneServer = newOcServiceCert("ca", constants.ServiceNameKeystone, constants.KeystoneCertName)
	// OcadmCertRegionServer is the definition of the cert used to serve compute controller service.
	OcadmCertRegionServer = newOcServiceCert("ca", constants.ServiceNameRegion, constants.RegionCertName)
	// OcadmCertClient is the definition of the cert used by the cli to access the api server
	OcadmClimcCertClient = OnecloudCert{
		Name:     "climc",
		LongName: "Client certificate for the console client to auth",
		BaseName: constants.ClimcClientCertAndKeyBaseName,
		CAName:   "ca",
		config: certutil.Config{
			CommonName: constants.ClimcClientCertAndKeyBaseName,
			Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
	}
)

func newOcServiceCert(caName string, serviceName string, certName string) OnecloudCert {
	return OnecloudCert{
		Name:     serviceName,
		LongName: fmt.Sprintf("certificate for serving the %s service", serviceName),
		BaseName: constants.KeystoneCertAndKeyBaseName,
		CAName:   caName,
		config: certutil.Config{
			CommonName: serviceName,
			Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		configMutators: []configMutatorsFunc{
			makeAltNamesMutator(
				pkiutil.GetServiceAltNames,
				serviceName,
				certName,
			),
		},
	}
}

func makeAltNamesMutator(
	f func(*apis.InitConfiguration, string, string) (*certutil.AltNames, error),
	serviceName, certName string,
) configMutatorsFunc {
	return func(mc *apis.InitConfiguration, cc *certutil.Config) error {
		altNames, err := f(mc, serviceName, certName)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}
