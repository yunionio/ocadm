package grafana

const GrafanaTempate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: monitor
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: grafana
  labels:
    app: grafana
subjects:
- kind: ServiceAccount
  name: grafana
  namespace: monitor
roleRef:
  kind: ClusterRole
  name: grafana
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  labels:
    app: grafana
  name: grafana
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "watch", "list"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: grafana
  name: grafana
  namespace: monitor
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitor
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      initContainers:
      - name: grafana-sc-datasources
        image: {{.SidecarImage}}
        imagePullPolicy: IfNotPresent
        env:
        - name: METHOD
          value: LIST
        - name: LABEL
          value: grafana_datasource
        - name: FOLDER
          value: "/etc/grafana/provisioning/datasources"
        - name: RESOURCE
          value: "both"
        volumeMounts:
        - name: sc-datasources-volume
          mountPath: "/etc/grafana/provisioning/datasources"
      containers:
      - name: grafana
        image: {{.Image}}
        imagePullPolicy: IfNotPresent
        env:
        - name: GF_SECURITY_ADMIN_USER
          valueFrom:
            secretKeyRef:
              key: admin-user
              name: grafana
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              key: admin-password
              name: grafana
        ports:
        - name: service
          containerPort: 80
          protocol: TCP
        - name: grafana
          containerPort: 3000
          protocol: TCP
        volumeMounts:
        - mountPath: /etc/grafana/grafana.ini
          name: config
          subPath: grafana.ini
        - mountPath: /etc/grafana/ldap.toml
          name: ldap
          subPath: ldap.toml
        - mountPath: /var/lib/grafana
          name: storage
        - mountPath: /etc/grafana/provisioning/datasources
          name: sc-datasources-volume
      serviceAccountName: grafana
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      volumes:
      - name: config
        configMap:
          defaultMode: 420
          name: grafana
      - name: ldap
        secret:
          defaultMode: 420
          items:
          - key: ldap-toml
            path: ldap.toml
          secretName: grafana
      - name: storage
        persistentVolumeClaim:
          claimName: grafana-data
      - name: sc-datasources-volume
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: grafana
  name: grafana
  namespace: monitor
spec:
  ports:
  - name: service
    port: 80
    protocol: TCP
    targetPort: 3000
  selector:
    app: grafana
  type: ClusterIP
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana
  namespace: monitor
  labels:
    app: grafana
data:
  grafana.ini: |
    [paths]
      data = /var/lib/grafana/data
      logs = /var/log/grafana
      plugins = /var/lib/grafana/plugins
      provisioning = /etc/grafana/provisioning
    [analytics]
      check_for_updates = true
    [log]
      mode = console
    [grafana_net]
      url = https://grafana.net
    [auth.anonymous]
      enabled = true
      org_role = Editor
    # [auth.ldap]
    #   enabled = true
    #   allow_sign_up = true
    #   config_file = /etc/grafana/ldap.toml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitor
  labels:
    app: grafana
    grafana_datasource: "true"
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - name: Loki
      type: loki
      isDefault: true
      access: proxy
      version: 1
      url: http://loki:3100
      jsonData:
        maxLines: 2000
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: grafana-data
  namespace: monitor
  labels:
    app: grafana
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10G
  storageClassName: local-path
  volumeMode: Filesystem
---
apiVersion: v1
kind: Secret
metadata:
  labels:
    app: grafana
  name: grafana
  namespace: monitor
type: Opaque
data:
  admin-password: YWRtaW5AMTIz
  admin-user: YWRtaW4=
  ldap-toml: ""
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  labels:
    app: grafana
  name: grafana
  namespace: monitor
spec:
  rules:
  - host: {{.IngressHost}}
    http:
      paths:
      - backend:
          serviceName: grafana
          servicePort: 80
---
`
