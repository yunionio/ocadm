package preflight

import (
	"fmt"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"net"
	"os"
	"yunion.io/x/ocadm/pkg/apis/v1"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	k8spreflight "k8s.io/kubernetes/cmd/kubeadm/app/preflight"
	utilruntime "k8s.io/kubernetes/cmd/kubeadm/app/util/runtime"
	utilsexec "k8s.io/utils/exec"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/util/mysql"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlCheck struct {
	*apis.MysqlConnection
}

func (MysqlCheck) Name() string {
	return "Mysql"
}

func (c MysqlCheck) Check() (warnings, errorList []error) {
	// 1. test connection
	// 2. check previllege
	conn, err := mysql.NewConnection(c.MysqlConnection)
	if err != nil {
		errorList = append(errorList, err)
		return
	}
	withGrant, err := conn.IsGrantPrivUser(c.Username, "%")
	if err != nil {
		errorList = append(errorList, err)
		return
	}
	if !withGrant {
		errorList = append(errorList, errors.Errorf("mysql user %s not 'WITH GRANT OPTION'", c.Username))
		return
	}

	return
}

// PortOpenCheck ensures the given port is available for use.
type PortOpenCheck struct {
	port  int
	label string
}

// Name returns name for PortOpenCheck. If not known, will return "PortXXXX" based on port number
func (poc PortOpenCheck) Name() string {
	if poc.label != "" {
		return poc.label
	}
	return fmt.Sprintf("Port-%d", poc.port)
}

// Check validates if the particular port is available.
func (poc PortOpenCheck) Check() (warnings, errorList []error) {
	klog.V(1).Infof("validating availability of port %d", poc.port)
	errorList = []error{}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", poc.port))
	if err != nil {
		errorList = append(errorList, errors.Errorf("Port %d is in use", poc.port))
	}
	if ln != nil {
		ln.Close()
	}

	return nil, errorList
}

// ImagePullCheck will pull container images used by ocadm and kubeadm
type ImagePullCheck struct {
	runtime   utilruntime.ContainerRuntime
	imageList []string
}

// Name returns the label for ImagePullCheck
func (ImagePullCheck) Name() string {
	return "ImagePull"
}

// Check pulls images required by ocadm and kubeadm. This is a mutating check
func (ipc ImagePullCheck) Check() (warnings, errorList []error) {
	for _, image := range ipc.imageList {
		ret, err := ipc.runtime.ImageExists(image)
		if ret && err == nil {
			klog.V(1).Infof("image exists: %s", image)
			continue
		}
		if err != nil {
			errorList = append(errorList, errors.Wrapf(err, "failed to check if image %s exists", image))
		}
		klog.V(1).Infof("pulling %s", image)
		if err := ipc.runtime.PullImage(image); err != nil {
			errorList = append(errorList, errors.Wrapf(err, "failed to pull image %s", image))
		}
	}
	return warnings, errorList
}

func RunInitNodeChecks(
	execer utilsexec.Interface,
	cfg *v1.InitConfiguration,
	kubeadmCfg *kubeadmapi.InitConfiguration,
	ignorePreflightErrors sets.String,
	isSecondaryControlPlane bool,
	downloadCerts bool,
) error {
	checks := []k8spreflight.Checker{
		MysqlCheck{
			MysqlConnection: &cfg.MysqlConnection,
		},
	}
	// Run onecloud preflight checks
	if err := k8spreflight.RunChecks(checks, os.Stderr, ignorePreflightErrors); err != nil {
		return err
	}
	// Run kubernetes preflight checks
	if err := k8spreflight.RunInitNodeChecks(
		execer,
		kubeadmCfg,
		ignorePreflightErrors,
		isSecondaryControlPlane,
		downloadCerts); err != nil {
		return errors.Wrap(err, "k8s init node checks")
	}
	return nil
}

func RunPullImagesCheck(execer utilsexec.Interface, cfg *v1.InitConfiguration, kubeadmCfg *kubeadmapi.InitConfiguration, ignorePreflightErrors sets.String) error {
	containerRuntime, err := utilruntime.NewContainerRuntime(utilsexec.New(), kubeadmCfg.NodeRegistration.CRISocket)
	if err != nil {
		return err
	}

	checks := []k8spreflight.Checker{
		ImagePullCheck{runtime: containerRuntime, imageList: images.GetControlPlaneImages(&kubeadmCfg.ClusterConfiguration)},
	}
	return k8spreflight.RunChecks(checks, os.Stderr, ignorePreflightErrors)
}
