package copycerts

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	//rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	keyutil "k8s.io/client-go/util/keyutil"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	nodebootstraptokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/copycerts"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	cryptoutil "k8s.io/kubernetes/cmd/kubeadm/app/util/crypto"

	"yunion.io/x/ocadm/pkg/apis/constants"
	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
)

var (
	CreateCertificateKey = copycerts.CreateCertificateKey
)

// createShortLivedBootstrapToken creates the token used to manager kubeadm-certs
// and return the tokenID
func createShortLivedBootstrapToken(client clientset.Interface) (string, error) {
	tokenStr, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return "", errors.Wrap(err, "error generating token to upload certs")
	}
	token, err := kubeadmapi.NewBootstrapTokenString(tokenStr)
	if err != nil {
		return "", errors.Wrap(err, "error creating upload certs token")
	}
	tokens := []kubeadmapi.BootstrapToken{{
		Token:       token,
		Description: "Proxy for managing TTL for the kubeadm-certs secret",
		TTL: &metav1.Duration{
			Duration: constants.DefaultCertTokenDuration,
		},
	}}

	if err := nodebootstraptokenphase.CreateNewTokens(client, tokens); err != nil {
		return "", errors.Wrap(err, "error creating token")
	}
	return tokens[0].Token.ID, nil
}

// UploadCerts save certs needs to create a new onecloud service
func UploadCerts(client clientset.Interface, cfg *apiv1.InitConfiguration, key string) error {
	fmt.Printf("[oc-upload-certs] Storing the certificates in ConfigMap %q in the %q namespace\n", constants.OcadmCertsSecret, metav1.NamespaceSystem)
	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	tokenID, err := createShortLivedBootstrapToken(client)
	if err != nil {
		return err
	}

	secretData, err := getDataFromDisk(cfg, decodedKey)
	if err != nil {
		return err
	}
	ref, err := getSecretOwnerRef(client, tokenID)
	if err != nil {
		return err
	}

	err = apiclient.CreateOrUpdateSecret(client, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            constants.OcadmCertsSecret,
			Namespace:       metav1.NamespaceSystem,
			OwnerReferences: ref,
		},
		Data: secretData,
	})
	if err != nil {
		return err
	}

	return nil
	//return createRBAC(client)
}

func getSecretOwnerRef(client clientset.Interface, tokenID string) ([]metav1.OwnerReference, error) {
	secretName := bootstraputil.BootstrapTokenSecretName(tokenID)
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error to get token reference")
	}

	gvk := schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
	ref := metav1.NewControllerRef(secret, gvk)
	return []metav1.OwnerReference{*ref}, nil
}

func loadAndEncryptCert(certPath string, key []byte) ([]byte, error) {
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	return cryptoutil.EncryptBytes(cert, key)
}

func certsToTransfer(cfg *apiv1.InitConfiguration) map[string]string {
	certsDir := cfg.OnecloudCertificatesDir
	certs := map[string]string{
		constants.CACertName: path.Join(certsDir, constants.CACertName),
		constants.CAKeyName:  path.Join(certsDir, constants.CAKeyName),
	}

	return certs
}

func getDataFromDisk(cfg *apiv1.InitConfiguration, key []byte) (map[string][]byte, error) {
	secretData := map[string][]byte{}
	for certName, certPath := range certsToTransfer(cfg) {
		cert, err := loadAndEncryptCert(certPath, key)
		if err == nil || (err != nil && os.IsNotExist(err)) {
			secretData[certOrKeyNameToSecretName(certName)] = cert
		} else {
			return nil, err
		}
	}
	return secretData, nil
}

// DownloadCerts downloads the certificates needed to join a new control plane.
func DownloadCerts(client clientset.Interface, cfg *apiv1.InitConfiguration, key string) error {
	fmt.Printf("[download-certs] Downloading the certificates in Secret %q in the %q Namespace\n", constants.OcadmCertsSecret, metav1.NamespaceSystem)

	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return errors.Wrap(err, "error decoding certificate key")
	}

	secret, err := getSecret(client)
	if err != nil {
		return errors.Wrap(err, "error downloading the secret")
	}

	secretData, err := getDataFromSecret(secret, decodedKey)
	if err != nil {
		return errors.Wrap(err, "error decoding secret data with provided key")
	}

	for certOrKeyName, certOrKeyPath := range certsToTransfer(cfg) {
		certOrKeyData, found := secretData[certOrKeyNameToSecretName(certOrKeyName)]
		if !found {
			return errors.New("couldn't find required certificate or key in Secret")
		}
		if len(certOrKeyData) == 0 {
			klog.V(1).Infof("[download-certs] Not saving %q to disk, since it is empty in the %q Secret\n", certOrKeyName, constants.OcadmCertsSecret)
			continue
		}
		if err := writeCertOrKey(certOrKeyPath, certOrKeyData); err != nil {
			return err
		}
	}

	return nil
}

func writeCertOrKey(certOrKeyPath string, certOrKeyData []byte) error {
	if _, err := keyutil.ParsePublicKeysPEM(certOrKeyData); err == nil {
		return keyutil.WriteKey(certOrKeyPath, certOrKeyData)
	} else if _, err := certutil.ParseCertsPEM(certOrKeyData); err == nil {
		return certutil.WriteCert(certOrKeyPath, certOrKeyData)
	}
	return errors.New("unknown data found in Secret entry")
}

func getSecret(client clientset.Interface) (*v1.Secret, error) {
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(constants.OcadmCertsSecret, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("Secret %q was not found in the %q Namespace. This Secret might have expired. Please, run `kubeadm init phase upload-certs --experimental-upload-certs` on a control plane to generate a new one", constants.OcadmCertsSecret, metav1.NamespaceSystem)
		}
		return nil, err
	}
	return secret, nil
}

func getDataFromSecret(secret *v1.Secret, key []byte) (map[string][]byte, error) {
	secretData := map[string][]byte{}
	for secretName, encryptedSecret := range secret.Data {
		// In some cases the secret might have empty data if the secrets were not present on disk
		// when uploading. This can specially happen with external insecure etcd (no certs)
		if len(encryptedSecret) > 0 {
			cert, err := cryptoutil.DecryptBytes(encryptedSecret, key)
			if err != nil {
				// If any of the decrypt operations fail do not return a partial result,
				// return an empty result immediately
				return map[string][]byte{}, err
			}
			secretData[secretName] = cert
		} else {
			secretData[secretName] = []byte{}
		}
	}
	return secretData, nil
}

func certOrKeyNameToSecretName(certOrKeyName string) string {
	return strings.Replace(certOrKeyName, "/", "-", -1)
}
