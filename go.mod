module yunion.io/x/ocadm

go 1.12

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/distribution v2.8.0+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20190109173153-a79fabbfe841 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libnetwork v0.8.0-dev.2.0.20200102182716-9fd385be8302 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mholt/caddy v1.0.0 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.19.3
	k8s.io/component-base v0.15.8
	k8s.io/klog v0.3.3
	k8s.io/kubernetes v1.16.0
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	sigs.k8s.io/cluster-api v0.1.4
	sigs.k8s.io/controller-runtime v0.1.11 // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	yunion.io/x/jsonutils v0.0.0-20211105163012-d846c05a3c9a
	yunion.io/x/log v0.0.0-20201210064738-43181789dc74
	yunion.io/x/onecloud v0.0.0-20211113105258-fa8a7d4a355f
	yunion.io/x/onecloud-operator v0.0.0-20211113111739-e27d781a1561
	yunion.io/x/pkg v0.0.0-20210918114143-ce839f862c5f
)

replace (
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	k8s.io/api => k8s.io/api v0.15.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.9-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.15.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.15.8
	k8s.io/client-go => github.com/zexi/client-go v0.15.9-rc.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.15.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.15.8
	k8s.io/code-generator => k8s.io/code-generator v0.15.9-beta.0
	k8s.io/component-base => k8s.io/component-base v0.15.8
	k8s.io/cri-api => k8s.io/cri-api v0.15.9-beta.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.15.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.15.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.15.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.15.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.15.8
	k8s.io/kubectl => k8s.io/kubectl v0.15.9-beta.0
	k8s.io/kubelet => k8s.io/kubelet v0.15.8
	k8s.io/kubernetes => github.com/zexi/kubernetes v1.15.9-rc.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.15.8
	k8s.io/metrics => k8s.io/metrics v0.15.8
	k8s.io/node-api => k8s.io/node-api v0.15.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.15.8
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.15.8
	k8s.io/sample-controller => k8s.io/sample-controller v0.15.8

)
