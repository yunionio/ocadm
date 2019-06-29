package pkiutil

import (
	"crypto"
	"crypto/rand"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	certutil "k8s.io/client-go/util/cert"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"k8s.io/kubernetes/pkg/registry/core/service/ipallocator"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
)

const (
	duration365d = time.Hour * 24 * 365
)

var (
	NewPrivateKey       = pkiutil.NewPrivateKey
	TryLoadCertFromDisk = pkiutil.TryLoadCertFromDisk
	TryLoadKeyFromDisk  = pkiutil.TryLoadKeyFromDisk
)

// NewCertAndKey creates new certificate and key by passing the certificate authority certificate and key
func NewCertAndKey(caCert *x509.Certificate, caKey *rsa.PrivateKey, config *certutil.Config) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create private key")
	}

	cert, err := NewSignedCert(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to sign certificate")
	}

	return cert, key, nil
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d * 10).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// GetServiceAltNames builds an AltNames object to be used when generating service certificate
func GetServiceAltNames(cfg *apis.InitConfiguration, serviceName string, certName string) (*certutil.AltNames, error) {
	// advertise address
	advertiseAddress := net.ParseIP(cfg.LocalAPIEndpoint.AdvertiseAddress)
	if advertiseAddress == nil {
		return nil, errors.Errorf("error parsing LocalAPIEndpoint AdvertiseAddress %v: is not a valid textual representation of an IP address",
			cfg.LocalAPIEndpoint.AdvertiseAddress)
	}

	// internal IP address for the API server
	_, svcSubnet, err := net.ParseCIDR(cfg.Networking.ServiceSubnet)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing CIDR %q", cfg.Networking.ServiceSubnet)
	}

	internalAPIServerVirtualIP, err := ipallocator.GetIndexedIP(svcSubnet, 1)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get first IP address from the given CIDR (%s)", svcSubnet.String())
	}

	// create AltNames with defaults DNSNames/IPs
	altNames := &certutil.AltNames{
		DNSNames: []string{
			cfg.NodeRegistration.Name,
			serviceName,
			fmt.Sprintf("%s.kube-system", serviceName),
			fmt.Sprintf("%s.kube-system.svc", serviceName),
			fmt.Sprintf("%s.kube-system.svc.%s", serviceName, cfg.Networking.DNSDomain),
		},
		IPs: []net.IP{
			internalAPIServerVirtualIP,
			advertiseAddress,
		},
	}

	// add cluster controlPlaneEndpoint if present (dns or ip)
	if len(cfg.ControlPlaneEndpoint) > 0 {
		if host, _, err := kubeadmutil.ParseHostPort(cfg.ControlPlaneEndpoint); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				altNames.IPs = append(altNames.IPs, ip)
			} else {
				altNames.DNSNames = append(altNames.DNSNames, host)
			}
		} else {
			return nil, errors.Wrapf(err, "error parsing cluster controlPlaneEndpoint %q", cfg.ControlPlaneEndpoint)
		}
	}

	appendSANsToAltNames(altNames, cfg.APIServer.CertSANs, certName)

	return altNames, nil
}

// appendSANsToAltNames parses SANs from as list of strings and adds them to altNames for use on a specific cert
// altNames is passed in with a pointer, and the struct is modified
// valid IP address strings are parsed and added to altNames.IPs as net.IP's
// RFC-1123 compliant DNS strings are added to altNames.DNSNames as strings
// certNames is used to print user facing warnings and should be the name of the cert the altNames will be used for
func appendSANsToAltNames(altNames *certutil.AltNames, SANs []string, certName string) {
	for _, altname := range SANs {
		if ip := net.ParseIP(altname); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else if len(validation.IsDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else {
			fmt.Printf(
				"[certificates] WARNING: '%s' was not added to the '%s' SAN, because it is not a valid IP or RFC-1123 compliant DNS entry\n",
				altname,
				certName,
			)
		}
	}
}
