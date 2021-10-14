# ocadm 作用

ocadm 用于部署在 Kubernetes 集群上运行的 Cloudpods 服务。

ocadm 包装了 kubeadm 的代码，用于部署 Kubernetes 集群，然后添加了额外的代码部署 cloudpods-operator，calico-cni 等关键服务，保证在部署完成的 Kubernetes 集群上运行 Cloudpods 服务。

# 编译

```bash
$ git clone https://github.com/yunionio/ocadm $GOPATH/src/yunion.io/x/ocadm
$ cd $GOPATH/src/yunion.io/x/ocadm
$ make
```

# 准备安装环境和依赖

## 环境

centos 7

## 安装配置依赖

1. mysql
```bash
$ MYSQL_PASSWD='your-sql-passwd'
$ yum install -y mariadb-server
$ systemctl enable mariadb
$ systemctl start mariadb
$ mysqladmin -u root password "$MYSQL_PASSWD"
$ cat <<EOF >/etc/my.cnf
[mysqld]
datadir=/var/lib/mysql
socket=/var/lib/mysql/mysql.sock
# Disabling symbolic-links is recommended to prevent assorted security risks
symbolic-links=0
# Settings user and group are ignored when systemd is used.
# If you need to run mysqld under a different user or group,
# customize your systemd unit file for mariadb according to the
# instructions in http://fedoraproject.org/wiki/Systemd
skip_name_resolve

[mysqld_safe]
log-error=/var/log/mariadb/mariadb.log
pid-file=/var/run/mariadb/mariadb.pid

#
# include all files from the config directory
#
!includedir /etc/my.cnf.d
EOF
$ mysql -uroot -p$MYSQL_PASSWD \
  -e "GRANT ALL ON *.* to 'root'@'%' IDENTIFIED BY '$MYSQL_PASSWD' with grant option; FLUSH PRIVILEGES;"
```

2. docker & kubelet
```bash
$ yum install -y yum-utils
$ yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
$ yum install -y docker-ce-18.09.1 docker-ce-cli-18.09.1 containerd.io
$ cat <<EOF >/etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=http://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg http://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
$ yum install --assumeyes bridge-utils conntrack-tools jq kubelet-1.14.3-0 kubeadm kubectl

$ mkdir /etc/docker
$ cat <<EOF >/etc/docker/daemon.json
{
  "bridge": "none",
  "iptables": false,
  "exec-opts":
    [
      "native.cgroupdriver=cgroupfs"
    ],
  "data-root": "/opt/docker",
  "live-restore": true,
  "log-driver": "json-file",
  "log-opts":
    {
      "max-size": "100m"
    },
  "registry-mirrors":
    [
      "https://lje6zxpk.mirror.aliyuncs.com",
      "https://lms7sxqp.mirror.aliyuncs.com",
      "https://registry.docker-cn.com"
    ],
  "storage-driver": "overlay2"
}
EOF
$ systemctl enable --now docker
$ systemctl enable kubelet
```

# init

```bash
# create kubernetes cluster
$ ./ocadm init --mysql-host "$MYSQL_HOST" --mysql-user root --mysql-password "$MYSQL_PASSWD"

# create onecloud cluster
$ ./ocadm cluster create

# get cluster
$ kubectl get onecloudcluster -n onecloud

# view cluster pods
$ kubectl get pods -n onecloud
```

# reset

```bash
$ ./ocadm reset
```

# configure onecloud component

```bash
$ ./ocadm node disable-host-agent          # Run this command to select node disable host agent
$ ./ocadm node disable-onecloud-controller # Run this command to select node disable onecloud controller
$ ./ocadm node enable-host-agent           # Run this command to select node enable host agent
$ ./ocadm node enable-onecloud-controller  # Run this command to select node enable onecloud controller
$ ./ocadm baremetal enable                 # Run this command to select node enable baremetal agent
$ ./ocadm baremetal disable                # Run this command to select node disable baremetal agent
```
