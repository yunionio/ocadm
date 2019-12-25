package onecloudoperator

const OperatorTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: onecloud-operator
subjects:
- kind: ServiceAccount
  namespace: {{.Namespace}}
  name: onecloud-operator
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: {{.Namespace}}
  name: onecloud-operator
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-app: onecloud-operator
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ''
  namespace: {{.Namespace}}
  name: onecloud-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: onecloud-operator
  template:
    metadata:
      labels:
        k8s-app: onecloud-operator
    spec:
      serviceAccount: onecloud-operator
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: node-role.kubernetes.io/controlplane
        effect: NoSchedule
      containers:
      - name: onecloud-operator
        image: {{.Image}}
        imagePullPolicy: Always
        command: ["/bin/onecloud-controller-manager"]
        env:
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
`

const LocalPathProvisioner = `
apiVersion: v1
kind: Namespace
metadata:
  name: local-path-storage
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: local-path-provisioner-service-account
  namespace: local-path-storage
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: local-path-provisioner-role
  namespace: local-path-storage
rules:
- apiGroups: [""]
  resources: ["nodes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["endpoints", "persistentvolumes", "pods"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: local-path-provisioner-bind
  namespace: local-path-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: local-path-provisioner-role
subjects:
- kind: ServiceAccount
  name: local-path-provisioner-service-account
  namespace: local-path-storage
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: local-path-provisioner
  namespace: local-path-storage
spec:
  replicas: 1
  selector:
    matchLabels:
      app: local-path-provisioner
  template:
    metadata:
      labels:
        app: local-path-provisioner
    spec:
      serviceAccountName: local-path-provisioner-service-account
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: node-role.kubernetes.io/controlplane
        effect: NoSchedule
      containers:
      - name: local-path-provisioner
        image: {{.Image}}
        imagePullPolicy: IfNotPresent
        command:
        - local-path-provisioner
        - --debug
        - start
        - --config
        - /etc/config/config.json
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config/
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      volumes:
        - name: config-volume
          configMap:
            name: local-path-config
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-path
provisioner: rancher.io/local-path
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: local-path-config
  namespace: local-path-storage
data:
  config.json: |-
        {
                "nodePathMap":[
                {
                        "node":"DEFAULT_PATH_FOR_NON_LISTED_NODES",
                        "paths":["/opt/local-path-provisioner"]
                }
                ]
        }

`

const TraefikTemplate = `
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
- kind: ServiceAccount
  name: traefik-ingress-controller
  namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      hostNetwork: true
      terminationGracePeriodSeconds: 60
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      - key: node-role.kubernetes.io/controlplane
        effect: NoSchedule
      containers:
      - image: {{.Image}}
        name: traefik-ingress-lb
        ports:
        - name: http
          containerPort: 80
        - name: admin
          containerPort: 8580
        securityContext:
          capabilities:
            drop:
            - ALL
            add:
            - NET_BIND_SERVICE
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
        - --defaultentrypoints=http,https
        - --entrypoints=Name:https Address::443 TLS
        - --entrypoints=Name:http Address::80 Redirect.EntryPoint:https
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
`
