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
apiVersion: scheduling.k8s.io/v1beta1
kind: PriorityClass
metadata:
  name: onecloud-operator-critical
value: 1000000000
globalDefault: false
description: "This priority class should be used for onecloud operator service pods only."
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
      priorityClassName: onecloud-operator-critical
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
