#!/bin/bash

set -e

GLOBALRC_FILE=${GLOBALRC_FILE:-/opt/cloud/workspace/globalrc}
VARS_FILE=${VARS_FILE:-/opt/yunionsetup/vars}

error_exit() {
    echo "Error: $1"
    exit 1
}

check_file_exists() {
    local file=$1
    if [ ! -f "$file" ]; then
        error_exit "file $file not exists"
    fi
}

new_cluster_yaml() {
    local rc_file=$GLOBALRC_FILE
    local vars_file=$VARS_FILE

    check_file_exists $rc_file
    check_file_exists $vars_file

    source $rc_file
    source $vars_file

    cat <<EOF
---
# onecloud cluster core components config
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-cluster-config
  namespace: onecloud
data:
  OnecloudClusterConfig: |
    apiVersion: onecloud.yunion.io/v1alpha1
    kind: OnecloudClusterConfig
    apiGateway:
      username: "$YUNIONAPI_ADMIN_USER"
      password: "$YUNIONAPI_ADMIN_PASS"
    glance:
      db:
        database: "$MYSQL_DB_GLANCE"
        username: "$MYSQL_USER_GLANCE"
        password: "$MYSQL_PASS_GLANCE"
      username: "$GLANCE_ADMIN_USER"
      password: "$GLANCE_ADMIN_PASS"
    keystone:
      db:
        database: "$MYSQL_DB_KEYSTONE"
        username: "$MYSQL_USER_KEYSTONE"
        password: "$MYSQL_PASS_KEYSTONE"
    kubeserver:
      db:
        database: "$MYSQL_DB_KUBE"
        username: "$MYSQL_USER_KUBE"
        password: "$MYSQL_PASS_KUBE"
      username: "$YUNION_KUBE_SERVER_ADMIN_USER"
      password: "$YUNION_KUBE_SERVER_ADMIN_PASS"
    logger:
      db:
        database: "$MYSQL_DB_LOGGER"
        username: "$MYSQL_USER_LOGGER"
        password: "$MYSQL_PASS_LOGGER"
      username: "$LOGGER_ADMIN_USER"
      password: "$LOGGER_ADMIN_PASS"
    region:
      db:
        database: "$MYSQL_DB_REGION"
        username: "$MYSQL_USER_REGION"
        password: "$MYSQL_PASS_REGION"
      username: "$REGION_ADMIN_USER"
      password: "$REGION_ADMIN_PASS"
    webconsole:
      username: "$YUNION_WEBCONSOLE_ADMIN_USER"
      password: "$YUNION_WEBCONSOLE_ADMIN_PASS"
    yunionagent:
      db:
        database: "$MYSQL_DB_YUNIONAGENT"
        username: "$MYSQL_USER_YUNIONAGENT"
        password: "$MYSQL_PASS_YUNIONAGENT"
      username: "$YUNIONAGENT_ADMIN_USER"
      password: "$YUNIONAGENT_ADMIN_PASS"
    yunionconf:
      db:
        database: "$MYSQL_DB_YUNIONCONF"
        username: "$MYSQL_USER_YUNIONCONF"
        password: "$MYSQL_PASS_YUNIONCONF"
      username: "$YUNIONCONF_ADMIN_USER"
      password: "$YUNIONCONF_ADMIN_PASS"
---
# onecloud cluster misc components config
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-cluster-components-config
  namespace: onecloud
data:
  OnecloudComponentsConfig: |
    cloudmon:
      username: "$YUNION_CLOUDMON_DOCKER_USER"
      password: "$YUNION_CLOUDMON_DOCKER_PSWD"
    cloudwatcher:
      db:
        database: "$MYSQL_DB_CLOUDWATCHER"
        username: "$MYSQL_USER_CLOUDWATCHER"
        password: "$MYSQL_PASS_CLOUDWATCHER"
      username: "$YUNION_CLOUDWATCHER_DOCKER_USER"
      password: "$YUNION_CLOUDWATCHER_DOCKER_PSWD"
    itsm:
      encryptionKey: "$ITSM_ENCRYPTION_KEY"
      secondDatabase: "$MYSQL_DB_ITSM_SECONDARY"
      db:
        database: "$MYSQL_DB_ITSM_PRIMARY"
        username: "$MYSQL_USER_ITSM_PRIMARY"
        password: "$MYSQL_PASS_ITSM"
      username: "$YUNION_ITSM_DOCKER_USER"
      password: "$YUNION_ITSM_DOCKER_PSWD"
    meter:
      db:
        database: "$MYSQL_DB_METER"
        username: "$MYSQL_USER_METER"
        password: "$MYSQL_PASS_METER"
      username: "$YUNION_METER_DOCKER_USER"
      password: "$YUNION_METER_DOCKER_PSWD"
    meteralert:
      db:
        database: "$MYSQL_DB_METERALERT"
        username: "$MYSQL_USER_METERALERT"
        password: "$MYSQL_PASS_METERALERT"
      username: "$YUNION_METERALERT_DOCKER_USER"
      password: "$YUNION_METERALERT_DOCKER_PSWD"
---
# default onecloud cluster
apiVersion: onecloud.yunion.io/v1alpha1
kind: OnecloudCluster
metadata:
  name: default
  namespace: onecloud
spec:
  loadBalancerEndpoint: "$MANAGEMENT_IP"
  imageRepository: registry.cn-beijing.aliyuncs.com/yunionio
  mysql:
    host: "$MYSQL_HOST"
    password: "$MYSQL_ROOT_PASSWORD"
    port: $MYSQL_PORT
    username: root
  region: "$REGION"
  zone: "$ZONE"
  keystone:
    bootstrapPassword: "$SYSADMIN_PASSWORD"
  glance:
    nodeSelector:
      kubernetes.io/hostname: glance-node
EOF
}

new_cluster_yaml
