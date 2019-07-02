# 编译

```bash
$ git clone https://github.com/zexi/ocadm $GOPATH/src/yunion.io/x/ocadm
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
  "bip": "172.17.0.1/16",
  "exec-opts":
    [
      "native.cgroupdriver=cgroupfs"
    ],
  "graph": "/opt/docker",
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

3. qemu

4. kernel

5. restart
重启使内核生效

# init

```bash
$ ./ocadm init --mysql-host <> --mysql-user root --mysql-password "$MYSQL_PASSWD" --kubernetes-version v1.14.3
```

# reset

```bash
$ ./ocadm reset
```
