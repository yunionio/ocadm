package options


import kubeadmoptions "k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"

var (
	// AddCertificateDirFlag adds the --certs-dir flag to the given flagset
	AddCertificateDirFlag = kubeadmoptions.AddCertificateDirFlag

	// AddCSRFlag adds the --csr-only flag to the given flagset
	AddCSRFlag = kubeadmoptions.AddCSRFlag

	// AddCSRDirFlag adds the --csr-dir flag to the given flagset
	AddCSRDirFlag  = kubeadmoptions.AddCSRDirFlag
)
