module yunion.io/x/ocadm

go 1.12

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libnetwork v0.0.0-20180830151422-a9cd636e3789 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mholt/caddy v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/vishvananda/netlink v0.0.0-20171020171820-b2de5d10e38e // indirect
	github.com/vishvananda/netns v0.0.0-20171111001504-be1fbeda1936 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.21.1 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.0.0-20190606204050-af9c91bd2759
	k8s.io/apiextensions-apiserver v0.0.0-20190606210616-f848dc7be4a4 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503 // indirect
	k8s.io/client-go v11.0.1-0.20190606204521-b8faab9c5193+incompatible
	k8s.io/cloud-provider v0.0.0-20190606212257-347f17c60af0 // indirect
	k8s.io/cluster-bootstrap v0.0.0-20190606212113-a4a4ceb6dbd9 // indirect
	k8s.io/code-generator v0.0.0-20190620073620-d55040311883
	k8s.io/component-base v0.0.0-20190606204607-bb6a29a90c31
	k8s.io/klog v0.3.3
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	k8s.io/kube-proxy v0.0.0-20190606211532-0764ecc02a7e // indirect
	k8s.io/kubelet v0.0.0-20190606211701-c7caf0079385 // indirect
	k8s.io/kubernetes v1.14.3
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	yunion.io/x/jsonutils v0.0.0-20190625054549-a964e1e8a051
	yunion.io/x/onecloud v0.0.0-20190626003349-5ad94587367b
	yunion.io/x/pkg v0.0.0-20190620104149-945c25821dbf
	yunion.io/x/structarg v0.0.0-20190625074850-3c0636a9fffe
)

replace (
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190620073620-d55040311883
)
