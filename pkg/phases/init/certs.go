package init

import (
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"k8s.io/kubernetes/pkg/util/normalizer"
	"strings"

	"github.com/pkg/errors"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	ocadmscheme "yunion.io/x/ocadm/pkg/apis/scheme"
	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	certsphase "yunion.io/x/ocadm/pkg/phases/certs"
)

var (
	genericLongDesc = normalizer.LongDesc(`
		Generates the %[1]s, and saves them into %[2]s.cert and %[2]s.key files.%[3]s

		If both files already exist, kubeadm skips the generation step and existing files will be used.
		` + cmdutil.AlphaDisclaimer)
)

var (
	csrOnly bool
	csrDir  string
)

// NewCertsPhase return the phase for the certs
func NewCertsPhase() workflow.Phase {
	return workflow.Phase{
		Name:   "oc-certs",
		Short:  "Onecloud certificate generation",
		Phases: newCertSubPhases(),
		Run:    runCerts,
		Long:   cmdutil.MacroCommandLongDescription,
	}
}

func localFlags() *pflag.FlagSet {
	set := pflag.NewFlagSet("csr", pflag.ExitOnError)
	options.AddCSRFlag(set, &csrOnly)
	options.AddCSRDirFlag(set, &csrDir)
	return set
}

// newCertSubPhases return sub phases for certs phase
func newCertSubPhases() []workflow.Phase {
	subPhases := []workflow.Phase{}

	// All subphase
	allPhase := workflow.Phase{
		Name:           "all",
		Short:          "Generates all certificates",
		InheritFlags:   getCertPhaseFlags("all"),
		RunAllSiblings: true,
	}

	subPhases = append(subPhases, allPhase)

	certTree, _ := certsphase.GetDefaultCertList().AsMap().CertTree()

	for ca, certList := range certTree {
		caPhase := newCertSubPhase(ca, runCAPhase(ca))
		subPhases = append(subPhases, caPhase)

		for _, cert := range certList {
			certPhase := newCertSubPhase(cert, runCertPhase(cert, ca))
			certPhase.LocalFlags = localFlags()
			subPhases = append(subPhases, certPhase)
		}
	}

	return subPhases
}

func newCertSubPhase(certSpec *certsphase.OnecloudCert, run func(c workflow.RunData) error) workflow.Phase {
	phase := workflow.Phase{
		Name:  certSpec.Name,
		Short: fmt.Sprintf("Generates the %s", certSpec.LongName),
		Long: fmt.Sprintf(
			genericLongDesc,
			certSpec.LongName,
			certSpec.BaseName,
			getSANDescription(certSpec),
		),
		Run:          run,
		InheritFlags: getCertPhaseFlags(certSpec.Name),
	}
	return phase
}

func getCertPhaseFlags(name string) []string {
	flags := []string{
		options.CertificatesDir,
		options.CfgPath,
		options.CSROnly,
		options.CSRDir,
	}
	/*if name == "all" || name == "apiserver" {
		flags = append(flags,
			options.APIServerAdvertiseAddress,
			options.APIServerCertSANs,
			options.NetworkingDNSDomain,
			options.NetworkingServiceSubnet,
		)
	}*/
	return flags
}

func getSANDescription(certSpec *certsphase.OnecloudCert) string {
	//Defaulted config we will use to get SAN certs
	defaultConfig := &apis.InitConfiguration{}
	ocadmscheme.Scheme.Default(defaultConfig)

	//Defaulted config we will use to get SAN certs
	defaultKubeadmConfig := &kubeadmapiv1beta1.InitConfiguration{
		LocalAPIEndpoint: kubeadmapiv1beta1.APIEndpoint{
			// GetAPIServerAltNames errors without an AdvertiseAddress; this is as good as any.
			AdvertiseAddress: "127.0.0.1",
		},
	}
	defaultKubeadmInternalConfig := &kubeadmapi.InitConfiguration{}
	kubeadmscheme.Scheme.Default(defaultKubeadmConfig)
	err := kubeadmscheme.Scheme.Convert(defaultKubeadmConfig, defaultKubeadmInternalConfig, nil)
	kubeadmutil.CheckErr(err)

	defaultConfig.InitConfiguration = *defaultKubeadmInternalConfig
	certConfig, err := certSpec.GetConfig(defaultConfig)
	kubeadmutil.CheckErr(err)

	if len(certConfig.AltNames.DNSNames) == 0 && len(certConfig.AltNames.IPs) == 0 {
		return ""
	}
	// This mutates the certConfig, but we're throwing it after we construct the command anyway
	sans := []string{}

	for _, dnsName := range certConfig.AltNames.DNSNames {
		if dnsName != "" {
			sans = append(sans, dnsName)
		}
	}

	for _, ip := range certConfig.AltNames.IPs {
		sans = append(sans, ip.String())
	}
	return fmt.Sprintf("\n\nDefault SANs are %s", strings.Join(sans, ", "))
}

func runCerts(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("certs phase invoked with an invalid data struct")
	}

	fmt.Printf("[oc-certs] Using certificateDir folder %q\n", data.OnecloudCertificateWriteDir())
	return nil
}

func runCAPhase(ca *certsphase.OnecloudCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(InitData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		if _, err := pkiutil.TryLoadCertFromDisk(data.OnecloudCertificateDir(), ca.BaseName); err == nil {
			if _, err := pkiutil.TryLoadKeyFromDisk(data.OnecloudCertificateDir(), ca.BaseName); err == nil {
				fmt.Printf("[certs] Using existing %s certificate authority\n", ca.BaseName)
				return nil
			}
			fmt.Printf("[certs] Using existing %s keyless certificate authority\n", ca.BaseName)
			return nil
		}

		// if dryrunning, write certificates authority to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.OnecloudCfg()
		cfg.CertificatesDir = data.OnecloudCertificateWriteDir()
		defer func() { cfg.CertificatesDir = data.OnecloudCertificateDir() }()

		// create the new certificate authority (or use existing)
		return certsphase.CreateCACertAndKeyFiles(ca, cfg)
	}
}

func runCertPhase(cert *certsphase.OnecloudCert, caCert *certsphase.OnecloudCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(InitData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		if certData, _, err := pkiutil.TryLoadCertAndKeyFromDisk(data.OnecloudCertificateDir(), cert.BaseName); err == nil {
			caCertData, err := pkiutil.TryLoadCertFromDisk(data.OnecloudCertificateDir(), caCert.BaseName)
			if err != nil {
				return errors.Wrapf(err, "couldn't load CA certificate %s", caCert.Name)
			}

			if err := certData.CheckSignatureFrom(caCertData); err != nil {
				return errors.Wrapf(err, "[oc-certs] certificate %s not signed by CA certificate %s", cert.BaseName, caCert.BaseName)
			}

			fmt.Printf("[oc-certs] Using existing %s certificate and key on disk\n", cert.BaseName)
			return nil
		}

		if csrOnly {
			fmt.Printf("[certs] Generating CSR for %s instead of certificate\n", cert.BaseName)
			if csrDir == "" {
				csrDir = data.OnecloudCertificateDir()
			}

			return certsphase.CreateCSR(cert, data.OnecloudCfg(), csrDir)
		}

		// if dryrunning, write certificates to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.OnecloudCfg()
		cfg.OnecloudCertificatesDir = data.OnecloudCertificateDir()
		defer func() { cfg.OnecloudCertificatesDir = data.OnecloudCertificateDir() }()

		// create the new certificate (or use existing)
		return certsphase.CreateCertAndKeyFilesWithCA(cert, caCert, cfg)
	}
}
