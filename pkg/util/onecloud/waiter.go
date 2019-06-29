package onecloud

import (
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"

	"yunion.io/x/ocadm/pkg/apis/constants"
)

type Waiter interface {
	apiclient.Waiter

	WaitForServicePods(serviceName string) error
	WaitForKeystone() error
	WaitForRegion() error
	WaitForGlance() error
}

type OCWaiter struct {
	apiclient.Waiter

	sessionFactory func() (*mcclient.ClientSession, error)
	timeout        time.Duration
	writer         io.Writer
}

// NewOCWaiter returns a new Onecloud waiter object that check service healthy
func NewOCWaiter(
	kubeClient clientset.Interface,
	sessionFactory func() (*mcclient.ClientSession, error),
	timeout time.Duration,
	writer io.Writer,
) Waiter {
	return &OCWaiter{
		Waiter:         apiclient.NewKubeWaiter(kubeClient, timeout, writer),
		sessionFactory: sessionFactory,
		timeout:        timeout,
		writer:         writer,
	}
}

func (w *OCWaiter) getSession() (*mcclient.ClientSession, error) {
	return w.sessionFactory()
}

func (w *OCWaiter) WaitForServicePods(serviceName string) error {
	if err := w.WaitForPodsWithLabel("component=" + serviceName); err != nil {
		return errors.Wrapf(err, "wait %s pod running", serviceName)
	}
	return nil
}

func (w *OCWaiter) WaitForKeystone() error {
	start := time.Now()
	return wait.PollImmediate(constants.APICallRetryInterval, w.timeout, func() (bool, error) {
		session, err := w.getSession()
		w.timeout.Seconds()
		if err != nil {
			duration := time.Since(start).Seconds()
			if (duration + float64(10*time.Second)) > w.timeout.Seconds() {
				fmt.Fprintf(w.writer, "[keystone] Error get auth session: %v", err)
			}
			return false, nil
		}
		if _, err := modules.Policies.List(session, nil); err != nil {
			return false, errors.Wrap(err, "Failed to get policy")
		}
		fmt.Printf("[keystone] healthy after %f seconds\n", time.Since(start).Seconds())
		return true, nil
	})
}

func (w *OCWaiter) waitForServiceHealthy(serviceName string, checkFunc func(*mcclient.ClientSession) (bool, error)) error {
	start := time.Now()
	return wait.PollImmediate(constants.APICallRetryInterval, w.timeout, func() (bool, error) {
		session, err := w.getSession()
		if err != nil {
			return false, errors.Errorf("Failed to get onecloud session: %v", session)
		}
		ok, err := checkFunc(session)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		fmt.Printf("[%s] healthy after %f seconds\n", serviceName, time.Since(start).Seconds())
		return true, nil
	})
}

func (w *OCWaiter) WaitForRegion() error {
	return w.waitForServiceHealthy(constants.ServiceNameRegionV2, func(s *mcclient.ClientSession) (bool, error) {
		_, err := modules.Servers.List(s, nil)
		if err == nil {
			return true, nil
		}
		return false, nil
	})
}

func (w *OCWaiter) WaitForGlance() error {
	return fmt.Errorf("not impl")
}
