package kube

import (
	"sync"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type Client struct {
	namespace  string
	config     genericclioptions.RESTClientGetter
	configOnce sync.Once

	KubeConfig  string
	KubeContext string
	Factory     cmdutil.Factory
}

func NewClient(kubeConfig string, namespace string, kubeContext string) (*Client, error) {
	c := &Client{
		namespace:   namespace,
		configOnce:  sync.Once{},
		KubeConfig:  kubeConfig,
		KubeContext: kubeContext,
	}
	getter := c.RESTClientGetter()
	// Add CRDs to the scheme. They are missing by default.
	if err := apiextv1beta1.AddToScheme(scheme.Scheme); err != nil {
		return c, err
	}
	c.Factory = cmdutil.NewFactory(getter)
	return c, nil
}

func NewClientByFile(configFile string) (*Client, error) {
	return NewClient(configFile, "", "")
}

// Namespace gets the namespace from the client
func (c *Client) Namespace() string {
	if c.namespace != "" {
		return c.namespace
	}
	if ns, _, err := c.RESTClientGetter().ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}

// RESTClientGetter gets the kubeconfig
func (c *Client) RESTClientGetter() genericclioptions.RESTClientGetter {
	c.configOnce.Do(func() {
		c.config = GetConfig(c.KubeConfig, c.KubeContext, c.namespace)
	})
	return c.config
}

func (c *Client) Rollout() (*Rollout, error) {
	return NewRollout(c)
}
