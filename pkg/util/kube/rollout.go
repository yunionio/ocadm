package kube

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/polymorphichelpers"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	"k8s.io/kubernetes/pkg/kubectl/util/interrupt"

	"yunion.io/x/log"
)

type Rollout struct {
	client *Client

	DynamicClient dynamic.Interface
}

func NewRollout(client *Client) (*Rollout, error) {
	f := client.Factory
	clientConfig, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	return &Rollout{
		client:        client,
		DynamicClient: dynamicClient,
	}, nil
}

func (r *Rollout) Status(timeout time.Duration) *RolloutStatus {
	status := &RolloutStatus{
		Rollout:        r,
		StatusViewerFn: polymorphichelpers.StatusViewerFn,
		Builder:        r.client.Factory.NewBuilder,
		Timeout:        timeout,
	}
	return status
}

type RolloutStatus struct {
	namespace string

	*Rollout

	StatusViewerFn func(*meta.RESTMapping) (kubectl.StatusViewer, error)
	Builder        func() *resource.Builder
	Timeout        time.Duration
	Revision       int64
}

func (r *RolloutStatus) SetNamespace(namespace string) *RolloutStatus {
	r.namespace = namespace
	return r
}

func (r *RolloutStatus) Namespace() string {
	if r.namespace == "" {
		return r.client.Namespace()
	}
	return r.namespace
}

func (r *RolloutStatus) RunDeployment(name string) error {
	return r.run("deployment", name)
}

func (r *RolloutStatus) RunDaemonset(name string) error {
	return r.run("daemonset", name)
}

func (r *RolloutStatus) RunStatefulset(name string) error {
	return r.run("statefulset", name)
}

func (r *RolloutStatus) run(resType string, name string) error {
	ret := r.Builder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(r.Namespace()).DefaultNamespace().
		ResourceTypeOrNameArgs(true, resType, name).
		SingleResourceType().
		Latest().
		Do()
	if err := ret.Err(); err != nil {
		return err
	}

	infos, err := ret.Infos()
	if err != nil {
		return err
	}
	if len(infos) != 1 {
		return errors.Errorf("rollout status is only supported on individual resources and resource collections - %d resources were found", len(infos))
	}
	info := infos[0]
	mapping := info.ResourceMapping()

	statusViewer, err := r.StatusViewerFn(mapping)
	if err != nil {
		return err
	}

	fieldSelector := fields.OneTermEqualSelector("metadata.name", info.Name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return r.DynamicClient.Resource(info.Mapping.Resource).Namespace(info.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return r.DynamicClient.Resource(info.Mapping.Resource).Namespace(info.Namespace).Watch(options)
		},
	}

	preconditionFunc := func(store cache.Store) (bool, error) {
		_, exists, err := store.Get(&metav1.ObjectMeta{Namespace: info.Namespace, Name: info.Name})
		if err != nil {
			return true, err
		}
		if !exists {
			// We need to make sure we see the object in the cache before we start waiting for events
			// or we would be waiting for the timeout if such object didn't exist.
			return true, apierrors.NewNotFound(mapping.Resource.GroupResource(), info.Namespace)
		}
		return false, nil
	}

	// if the rollout isn't done yet, keep watching deployment status
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), r.Timeout)
	intr := interrupt.New(nil, cancel)
	return intr.Run(func() error {
		_, err = watchtools.UntilWithSync(ctx, lw, &unstructured.Unstructured{}, preconditionFunc, func(e watch.Event) (bool, error) {
			switch t := e.Type; t {
			case watch.Added, watch.Modified:
				status, done, err := statusViewer.Status(e.Object.(runtime.Unstructured), r.Revision)
				if err != nil {
					return false, err
				}
				log.Infof("%s/%s status: %v", resType, name, status)
				// Quit waiting if the rollout is done
				if done {
					return true, nil
				}
				return false, nil
			case watch.Deleted:
				// We need to abort to avoid cases of recreation and not to silently watch the wrong (new) object
				return true, errors.Errorf("object has been deleted")
			default:
				return true, errors.Errorf("internal error: unexpected event %#v", e)
			}
		})
		return err
	})
}
