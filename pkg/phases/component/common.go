package component

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	onecloud "yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/controller"
	"yunion.io/x/onecloud-operator/pkg/label"
	"yunion.io/x/onecloud-operator/pkg/manager/component"
	onecloudutil "yunion.io/x/onecloud-operator/pkg/util/onecloud"
	"yunion.io/x/onecloud/pkg/mcclient"
)

const (
	JAVA_APP_JAR         = "JAVA_APP_JAR"
	JAVA_APP_WORKING_DIR = "/deployments"
)

var (
	GetOwnerRef      = controller.GetOwnerRef
	GetComponentName = controller.NewClusterComponentName
)

func GetLabel(oc *onecloud.OnecloudCluster, componentType onecloud.ComponentType) label.Label {
	instanceName := oc.GetLabels()[label.InstanceLabelKey]
	return label.New().Instance(instanceName).Component(componentType.String())
}

func GetObjectMeta(oc *onecloud.OnecloudCluster, name string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            name,
		Namespace:       oc.GetNamespace(),
		Labels:          labels,
		OwnerReferences: []metav1.OwnerReference{GetOwnerRef(oc)},
	}
}

func NewDeployment(
	cType onecloud.ComponentType,
	oc *onecloud.OnecloudCluster,
	volHelper *component.VolumeHelper,
	initContainersF func([]corev1.VolumeMount) []corev1.Container,
	containersF func([]corev1.VolumeMount) []corev1.Container,
	hostNetwork bool,
	dnsPolicy corev1.DNSPolicy,
) (*apps.Deployment, error) {
	labels := GetLabel(oc, cType)
	deployName := GetComponentName(oc.GetName(), cType)

	vols := volHelper.GetVolumes()
	volMounts := volHelper.GetVolumeMounts()

	var r1 int32 = 1
	appDeploy := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deployName,
			Namespace:       oc.GetNamespace(),
			Labels:          labels.Labels(),
			OwnerReferences: []metav1.OwnerReference{GetOwnerRef(oc)},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &r1,
			Strategy: apps.DeploymentStrategy{Type: apps.RecreateDeploymentStrategyType},
			Selector: labels.LabelSelector(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels.Labels(),
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/controlplane",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
					Containers:    containersF(volMounts),
					RestartPolicy: corev1.RestartPolicyAlways,
					Volumes:       vols,
					HostNetwork:   hostNetwork,
					DNSPolicy:     dnsPolicy,
				},
			},
		},
	}
	if initContainersF != nil {
		appDeploy.Spec.Template.Spec.InitContainers = initContainersF(volMounts)
	}
	return appDeploy, nil
}

func NewDefaultDeployment(
	cType onecloud.ComponentType,
	oc *onecloud.OnecloudCluster,
	volHelper *component.VolumeHelper,
	containersF func([]corev1.VolumeMount) []corev1.Container,
) (*apps.Deployment, error) {
	return NewDeployment(cType, oc, volHelper, nil, containersF, false, corev1.DNSClusterFirst)
}

func NewDefaultDeploymentWithHostNetwork(
	cType onecloud.ComponentType,
	oc *onecloud.OnecloudCluster,
	volHelper *component.VolumeHelper,
	containersF func([]corev1.VolumeMount) []corev1.Container,
) (*apps.Deployment, error) {
	return NewDeployment(cType, oc, volHelper, nil, containersF, true, corev1.DNSClusterFirstWithHostNet)
}

func NewService(
	cType onecloud.ComponentType,
	oc *onecloud.OnecloudCluster,
	serviceType corev1.ServiceType,
	ports []corev1.ServicePort,
) *corev1.Service {
	svcName := controller.NewClusterComponentName(oc.GetName(), cType)
	appLabel := GetLabel(oc, cType)
	return &corev1.Service{
		ObjectMeta: GetObjectMeta(oc, svcName, appLabel),
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Selector: appLabel,
			Ports:    ports,
		},
	}
}

func NewSinglePortService(cType onecloud.ComponentType, oc *onecloud.OnecloudCluster, port int32) *corev1.Service {
	ports := []corev1.ServicePort{
		component.NewServiceNodePort("api", port),
	}
	return NewService(cType, oc, corev1.ServiceTypeClusterIP, ports)
}

func NewNodePortService(cType onecloud.ComponentType, oc *onecloud.OnecloudCluster, port int32) *corev1.Service {
	ports := []corev1.ServicePort{
		component.NewServiceNodePort("api", port),
	}
	return NewService(cType, oc, corev1.ServiceTypeNodePort, ports)
}

func NewConfigMap(cType onecloud.ComponentType, oc *onecloud.OnecloudCluster, config string) *corev1.ConfigMap {
	name := controller.ComponentConfigMapName(oc, cType)
	return &corev1.ConfigMap{
		ObjectMeta: GetObjectMeta(oc, name, GetLabel(oc, cType).Labels()),
		Data: map[string]string{
			"config": config,
		},
	}
}

func NewConfigMapByTemplate(cType onecloud.ComponentType, oc *onecloud.OnecloudCluster, template string, config interface{}) (*corev1.ConfigMap, error) {
	data, err := component.CompileTemplateFromMap(template, config)
	if err != nil {
		return nil, err
	}
	return NewConfigMap(cType, oc, data), nil
}

func encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func SetObjectLastAppliedConfigAnnotation(
	obj metav1.Object,
	getSpecF func(metav1.Object) interface{},
) error {
	if getSpecF == nil {
		return nil
	}
	specApply, err := encode(getSpecF(obj))
	if err != nil {
		return err
	}
	anno := obj.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}
	anno[component.LastAppliedConfigAnnotation] = specApply
	obj.SetAnnotations(anno)
	return nil
}

type ComponentManager struct {
	kubeCli kubernetes.Interface
	session *mcclient.ClientSession
	cfg     *OnecloudComponentsConfig
}

func NewComponentManager(kubeCli kubernetes.Interface, session *mcclient.ClientSession, cfg *OnecloudComponentsConfig) *ComponentManager {
	return &ComponentManager{
		kubeCli: kubeCli,
		session: session,
		cfg:     cfg,
	}
}

func (m *ComponentManager) GetComponentsConfig() *OnecloudComponentsConfig {
	return m.cfg
}

func (m *ComponentManager) GetCloudSession() *mcclient.ClientSession {
	return m.session
}

func (m *ComponentManager) SyncComponent(oc *onecloud.OnecloudCluster, comp IComponent) error {
	if err := m.SyncService(oc, comp.NewService); err != nil {
		return err
	}
	if err := m.SyncConfigMap(oc, comp.NewDBConfig, comp.NewDBConfig2, comp.NewCloudUser, comp.NewConfigMap); err != nil {
		return err
	}
	if err := m.SyncDeployment(oc, comp.NewDeployment); err != nil {
		return err
	}
	if err := SyncCloudEndpoint(oc, comp.GetComponentType(), m.GetCloudSession(), comp.NewCloudEndpoint()); err != nil {
		return err
	}
	return nil
}

func (m *ComponentManager) SyncService(
	oc *onecloud.OnecloudCluster,
	svcFactory func(*onecloud.OnecloudCluster) *corev1.Service,
) error {
	newSvc := svcFactory(oc)
	if newSvc == nil {
		return nil
	}
	ns := oc.GetNamespace()
	isExistsF := func(obj metav1.Object) (metav1.Object, error) {
		svc := obj.(*corev1.Service)
		return m.kubeCli.CoreV1().Services(ns).Get(svc.GetName(), metav1.GetOptions{})
	}
	createF := func(obj metav1.Object) error {
		svc := obj.(*corev1.Service)
		_, err := m.kubeCli.CoreV1().Services(ns).Create(svc)
		return err
	}
	getSpecF := func(obj metav1.Object) interface{} {
		return obj.(*corev1.Service).Spec
	}
	equalF := equalFactory(&corev1.ServiceSpec{}, getSpecF)
	updateF := func(newObj, oldObj metav1.Object) error {
		newSvc := newObj.(*corev1.Service)
		oldSvc := oldObj.(*corev1.Service)
		svc := *oldSvc
		svc.Spec = newSvc.Spec
		svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		if err := SetObjectLastAppliedConfigAnnotation(&svc, getSpecF); err != nil {
			return err
		}
		if _, err := m.kubeCli.CoreV1().Services(ns).Update(&svc); err != nil {
			return err
		}
		return nil
	}
	return SyncK8sResource(oc, newSvc, isExistsF, createF, getSpecF, equalF, updateF)
}

func (m *ComponentManager) SyncConfigMap(
	oc *onecloud.OnecloudCluster,
	dbConfigF, dbConfigF2 func(*OnecloudComponentsConfig) *onecloud.DBConfig,
	cloudUserF func(*OnecloudComponentsConfig) *onecloud.CloudUser,
	cfgMapF func(*onecloud.OnecloudCluster, *OnecloudComponentsConfig) (*corev1.ConfigMap, error),
) error {
	clusterCfg := m.GetComponentsConfig()
	if dbConfigF != nil {
		for _, f := range []func(*OnecloudComponentsConfig) *onecloud.DBConfig{dbConfigF, dbConfigF2} {
			dbConfig := f(clusterCfg)
			if dbConfig != nil {
				if err := component.EnsureClusterDBUser(oc, *dbConfig); err != nil {
					return err
				}
			}
		}
	}
	if cloudUserF != nil {
		account := cloudUserF(clusterCfg)
		if account != nil {
			s := m.GetCloudSession()
			if err := component.EnsureServiceAccount(s, *account); err != nil {
				return err
			}
		}
	}
	cfgMap, err := cfgMapF(oc, clusterCfg)
	if err != nil {
		return err
	}
	if cfgMap == nil {
		return nil
	}
	return SyncConfigMap(m.kubeCli, oc, cfgMap)
}

func (m *ComponentManager) SyncDeployment(
	oc *onecloud.OnecloudCluster,
	deployF func(*onecloud.OnecloudCluster) (*apps.Deployment, error),
) error {
	newDeploy, err := deployF(oc)
	if err != nil {
		return err
	}
	if newDeploy == nil {
		return nil
	}
	ns := oc.GetNamespace()
	isExistsF := func(obj metav1.Object) (metav1.Object, error) {
		deploy := obj.(*apps.Deployment)
		return m.kubeCli.AppsV1().Deployments(ns).Get(deploy.GetName(), metav1.GetOptions{})
	}
	createF := func(obj metav1.Object) error {
		deploy := obj.(*apps.Deployment)
		_, err := m.kubeCli.AppsV1().Deployments(ns).Create(deploy)
		return err
	}
	getSpecF := func(obj metav1.Object) interface{} {
		return obj.(*apps.Deployment).Spec
	}
	equalF := func(newObj, oldObj metav1.Object) (bool, error) {
		oldConfig := apps.DeploymentSpec{}
		new := newObj.(*apps.Deployment)
		old := oldObj.(*apps.Deployment)
		if lastAppliedConfig, ok := old.Annotations[component.LastAppliedConfigAnnotation]; ok {
			err := json.Unmarshal([]byte(lastAppliedConfig), &oldConfig)
			if err != nil {
				return false, err
			}
			return apiequality.Semantic.DeepEqual(oldConfig.Replicas, new.Spec.Replicas) &&
				apiequality.Semantic.DeepEqual(oldConfig.Template, new.Spec.Template) &&
				apiequality.Semantic.DeepEqual(oldConfig.Strategy, new.Spec.Strategy), nil
		}
		return false, nil
	}
	updateF := func(newObj, oldObj metav1.Object) error {
		newDeploy := newObj.(*apps.Deployment)
		oldDeploy := oldObj.(*apps.Deployment)
		deploy := *oldDeploy
		deploy.Spec.Template = newDeploy.Spec.Template
		*deploy.Spec.Replicas = *newDeploy.Spec.Replicas
		deploy.Spec.Strategy = newDeploy.Spec.Strategy
		_, err := m.kubeCli.AppsV1().Deployments(ns).Update(&deploy)
		return err
	}
	return SyncK8sResource(oc, newDeploy, isExistsF, createF, getSpecF, equalF, updateF)
}

func equalFactory(oldSpec interface{}, getSpec func(metav1.Object) interface{}) func(newObj, oldObj metav1.Object) (bool, error) {
	return func(newObj, oldObj metav1.Object) (bool, error) {
		if lastAppliedConfig, ok := oldObj.GetAnnotations()[component.LastAppliedConfigAnnotation]; ok {
			err := json.Unmarshal([]byte(lastAppliedConfig), oldSpec)
			if err != nil {
				return false, errors.Wrapf(err, "unmarshal spec: [%s/%s]'s applied config failed", oldObj.GetNamespace(), oldObj.GetName())
			}
			return apiequality.Semantic.DeepEqual(oldSpec, getSpec(newObj)), nil
		}
		return false, nil
	}
}

func SyncK8sResource(
	oc *onecloud.OnecloudCluster,
	newObj metav1.Object,
	isExists func(obj metav1.Object) (metav1.Object, error),
	createF func(obj metav1.Object) error,
	getSpecF func(obj metav1.Object) interface{},
	equalF func(newObj, oldObj metav1.Object) (bool, error),
	updateF func(newObj, oldObj metav1.Object) error,
) error {
	oldObj, err := isExists(newObj)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if apierrors.IsNotFound(err) {
		if err := SetObjectLastAppliedConfigAnnotation(newObj, getSpecF); err != nil {
			return err
		}
		return createF(newObj)
	}

	equal, err := equalF(newObj, oldObj)
	if err != nil {
		return err
	}
	if !equal {
		return updateF(newObj, oldObj)
	}
	return nil
}

func GetImage(oc *onecloud.OnecloudCluster, componentType onecloud.ComponentType, version string) string {
	if version == "" {
		version = oc.Spec.Version
	}
	return fmt.Sprintf("%s/%s:%s", oc.Spec.ImageRepository, componentType.String(), version)
}

func GetJavaAppImage(oc *onecloud.OnecloudCluster, version string) string {
	return GetImage(oc, "java-app", version)
}

func SetJavaConfigVolumeMounts(volMounts []corev1.VolumeMount) []corev1.VolumeMount {
	confVol := volMounts[len(volMounts)-1]
	confVol.MountPath = fmt.Sprintf("%s/config", JAVA_APP_WORKING_DIR)
	volMounts[len(volMounts)-1] = confVol
	return volMounts
}

func SetJavaConfigVolumes(vols []corev1.Volume) []corev1.Volume {
	config := vols[len(vols)-1]
	config.ConfigMap.Items[0].Path = "application.properties"
	vols[len(vols)-1] = config
	return vols
}

func NewVolumeHelper(oc *onecloud.OnecloudCluster, cType onecloud.ComponentType) *component.VolumeHelper {
	return component.NewVolumeHelper(oc, controller.ComponentConfigMapName(oc, cType), cType)
}

func SyncConfigMap(
	cli kubernetes.Interface,
	oc *onecloud.OnecloudCluster,
	cfgMap *corev1.ConfigMap) error {
	ns := oc.GetNamespace()
	isExistsF := func(obj metav1.Object) (metav1.Object, error) {
		cfg := obj.(*corev1.ConfigMap)
		return cli.CoreV1().ConfigMaps(ns).Get(cfg.GetName(), metav1.GetOptions{})
	}
	createF := func(obj metav1.Object) error {
		cfg := obj.(*corev1.ConfigMap)
		_, err := cli.CoreV1().ConfigMaps(oc.GetNamespace()).Create(cfg)
		return err
	}
	equalF := func(_, _ metav1.Object) (bool, error) {
		return false, nil
	}
	updateF := func(newObj, oldObj metav1.Object) error {
		newCfg := newObj.(*corev1.ConfigMap)
		_, err := cli.CoreV1().ConfigMaps(ns).Update(newCfg)
		return err
	}
	return SyncK8sResource(oc, cfgMap, isExistsF, createF, nil, equalF, updateF)
}

func (m *ComponentManager) DisableComponent(oc *onecloud.OnecloudCluster, comp IComponent) error {
	if ep := comp.NewCloudEndpoint(); ep != nil {
		if err := DeleteCloudEndpoint(m.GetCloudSession(), ep); err != nil {
			return err
		}
	}
	if err := m.DeleteDeployment(oc, GetComponentName(oc.GetName(), comp.GetComponentType())); err != nil {
		return err
	}
	return nil
}

func (m *ComponentManager) DeleteDeployment(
	oc *onecloud.OnecloudCluster,
	name string,
) error {
	deleteF := func(name string) error {
		return m.kubeCli.AppsV1().Deployments(oc.GetNamespace()).Delete(name, &metav1.DeleteOptions{})
	}
	return DeleteK8sResource(name, deleteF)
}

func SyncCloudEndpoint(oc *onecloud.OnecloudCluster, cType onecloud.ComponentType, s *mcclient.ClientSession, ep *CloudEndpoint) error {
	if ep == nil {
		return nil
	}

	internalAddress := GetComponentName(oc.GetName(), cType)
	publicAddress := oc.Spec.LoadBalancerEndpoint
	if publicAddress == "" {
		publicAddress = internalAddress
	}

	urls := map[string]string{
		constants.EndpointTypePublic:   ep.GetUrl(publicAddress),
		constants.EndpointTypeInternal: ep.GetUrl(internalAddress),
	}
	return onecloudutil.RegisterServiceEndpoints(s, oc.Spec.Region, ep.ServiceName, ep.ServiceType, urls)
}

func DeleteCloudEndpoint(s *mcclient.ClientSession, ep *CloudEndpoint) error {
	if ep == nil {
		return nil
	}
	return onecloudutil.DeleteServiceEndpoints(s, ep.ServiceName)
}

func DeleteK8sResource(name string, deleteF func(name string) error) error {
	if err := deleteF(name); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}
