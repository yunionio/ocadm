package scheme

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"yunion.io/x/ocadm/pkg/apis/v1"
)

// Scheme is the runtime.Scheme to which all deployer api types are registered.
var Scheme = runtime.NewScheme()

// Codecs provides access to encoding and decoding for the scheme.
var Codecs = serializer.NewCodecFactory(Scheme)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	AddToScheme(Scheme)
}

// AddToScheme builds the deployer scheme using all knowns version of the deployer api.
func AddToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(v1.AddtoScheme(scheme))
}
