module yunion.io/x/ocadm

go 1.12

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/lithammer/dedent v1.1.0
	github.com/mholt/caddy v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v0.3.3
	k8s.io/kubernetes v1.15.1
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	sigs.k8s.io/cluster-api v0.1.4
	sigs.k8s.io/controller-runtime v0.1.11 // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	yunion.io/x/jsonutils v0.0.0-20191005115334-bb1c187fc0e7
	yunion.io/x/log v0.0.0-20190629062853-9f6483a7103d
	yunion.io/x/onecloud v0.0.0-20191210025243-fdf6f1cbdefd
	yunion.io/x/onecloud-operator v0.0.1-alpha3.0.20191231092239-6c398b6a9bf8
	yunion.io/x/pkg v0.0.0-20191121110824-e03b47b93fe0
)

replace (
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	k8s.io/api => k8s.io/api v0.0.0-20190718183219-b59d8169aab5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190718184206-a1aa83af71a7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190718185405-0ce9869d0015
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190718190308-f8e43aa19282
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190718190146-f7b0473036f9
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190718183727-0ececfbe9772
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190718190424-bef8d46b95de
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190718184434-a064d4d1ed7a
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190718190030-ea930fedc880
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190718185641-5233cb7cb41e
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190718185913-d5429d807831
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190718190949-4b42db8df903
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190718185757-9b45f80d5747
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190718190548-039b99e58dbd
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190718185242-1e1642704fe6
	k8s.io/node-api => k8s.io/node-api v0.0.0-20190718190710-3ae13b6d96d5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190718184639-baafa86838c0
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20190718185524-98384bc7a19f
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20190718184820-732eab031d75
)
