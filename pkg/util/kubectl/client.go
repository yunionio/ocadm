package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	tcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/clientcmd"
	"sigs.k8s.io/cluster-api/pkg/util"
)

const (
	retryIntervalKubectlApply = 10 * time.Second
	timeoutKubectlApply       = 15 * time.Minute
	retryIntervalKubectlLabel = 5 * time.Second
	timeoutKubectlLabel       = 5 * time.Minute
)

type Client struct {
	kubeconfigFile  string
	configOverrides tcmd.ConfigOverrides
	closeFn         func() error
}

func NewClientFormKubeconfigFile(filePath string) (*Client, error) {
	return &Client{
		kubeconfigFile:  filePath,
		configOverrides: clientcmd.NewConfigOverrides(),
	}, nil
}

func NewClientFromKubeconfig(kubeconfig string) (*Client, error) {
	f, err := createTempFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	defer ifErrRemove(err, f)
	c := &Client{
		kubeconfigFile:  f,
		configOverrides: clientcmd.NewConfigOverrides(),
	}
	c.closeFn = c.removeKubeconfigFile
	return c, nil
}

func createTempFile(contents string) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer ifErrRemove(err, f.Name())
	if err = f.Close(); err != nil {
		return "", err
	}
	err = ioutil.WriteFile(f.Name(), []byte(contents), 0644)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func ifErrRemove(err error, path string) {
	if err != nil {
		if err := os.Remove(path); err != nil {
			klog.Warningf("Error removing file '%s': %v", path, err)
		}
	}
}

func (c *Client) removeKubeconfigFile() error {
	return os.Remove(c.kubeconfigFile)
}

func (c *Client) kubectlManifestCmd(commandName, manifest string) error {
	cmd := exec.Command("kubectl", c.buildKubectlArgs(commandName)...)
	cmd.Stdin = strings.NewReader(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't kubectl apply, output: %s, error: %v", string(output), err)
	}
	return nil
}

func (c *Client) buildKubectlArgs(commandName string) []string {
	args := []string{commandName}
	if c.kubeconfigFile != "" {
		args = append(args, "--kubeconfig", c.kubeconfigFile)
	}
	if c.configOverrides.Context.Cluster != "" {
		args = append(args, "--cluster", c.configOverrides.Context.Cluster)
	}
	if c.configOverrides.Context.Namespace != "" {
		args = append(args, "--namespace", c.configOverrides.Context.Namespace)
	}
	if c.configOverrides.Context.AuthInfo != "" {
		args = append(args, "--user", c.configOverrides.Context.AuthInfo)
	}
	return append(args, "-f", "-")
}

func (c *Client) Apply(manifest string) error {
	return c.waitForKubectlApply(manifest)
}

func (c *Client) kubectlDelete(manifest string) error {
	return c.kubectlManifestCmd("delete", manifest)
}

func (c *Client) kubectlApply(manifest string) error {
	return c.kubectlManifestCmd("apply", manifest)
}

func (c *Client) waitForKubectlApply(manifest string) error {
	err := util.PollImmediate(retryIntervalKubectlApply, timeoutKubectlApply, func() (bool, error) {
		klog.V(1).Infof("Waiting for kubectl apply...")
		err := c.kubectlApply(manifest)
		if err != nil {
			if strings.Contains(err.Error(), "refused") {
				// Connection was refused, probably because the API server is not ready yet.
				klog.Infof("aiting for kubectl apply... server not yet available: %v", err)
				return false, nil
			}
			if strings.Contains(err.Error(), "unable to recognize") {
				klog.Infof("Waiting for kubectl apply... api not yet available: %v", err)
				return false, nil
			}
			klog.Warningf("Waiting for kubectl apply... unknown error %v", err)
			return false, err
		}

		return true, nil
	})
	return err
}

func (c *Client) Label(resource, resourceName, label string) error {
	return c.waitForKubectlLabel(resource, resourceName, label)
}

func (c *Client) waitForKubectlLabel(resource, resourceName, label string) error {
	return util.PollImmediate(retryIntervalKubectlLabel, timeoutKubectlLabel, func() (bool, error) {
		klog.V(1).Infof("Waiting for kubectl label")
		err := c.kubectlLabel(resource, resourceName, label)
		if err != nil {
			klog.Warning("Waiting for kubectl label error %s", err)
			return false, err
		}
		return true, nil
	})
}

func (c *Client) kubectlLabel(resource, resourceName, label string) error {
	output, err := exec.Command("kubectl", "label", resource, resourceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't kubectl label, output: %s, error: %v", string(output), err)
	}
	return nil
}
