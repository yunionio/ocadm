// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component

import (
	"path"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"yunion.io/x/onecloud/pkg/monitor/options"

	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	"yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"
	"yunion.io/x/onecloud-operator/pkg/controller"
	"yunion.io/x/onecloud-operator/pkg/manager"
)

type monitorManager struct {
	*ComponentManager
}

func newMonitorManager(man *ComponentManager) manager.Manager {
	return &monitorManager{man}
}

func (m *monitorManager) Sync(oc *v1alpha1.OnecloudCluster) error {
	return syncComponent(m, oc, oc.Spec.Monitor.Disable)
}

func (m *monitorManager) getDBConfig(cfg *v1alpha1.OnecloudClusterConfig) *v1alpha1.DBConfig {
	return &cfg.Monitor.DB
}

func (m *monitorManager) getCloudUser(cfg *v1alpha1.OnecloudClusterConfig) *v1alpha1.CloudUser {
	return &cfg.Monitor.CloudUser
}

func (m *monitorManager) getPhaseControl(man controller.ComponentManager) controller.PhaseControl {
	return controller.NewRegisterEndpointComponent(man, v1alpha1.MonitorComponentType,
		constants.ServiceNameMonitor, constants.ServiceTypeMonitor,
		constants.MonitorPort, "")
}

func (m *monitorManager) getConfigMap(oc *v1alpha1.OnecloudCluster, cfg *v1alpha1.OnecloudClusterConfig) (*corev1.ConfigMap, error) {
	opt := &options.Options
	if err := SetOptionsDefault(opt, constants.ServiceTypeMonitor); err != nil {
		return nil, err
	}
	config := cfg.Monitor
	SetDBOptions(&opt.DBOptions, oc.Spec.Mysql, config.DB)
	SetOptionsServiceTLS(&opt.BaseOptions)
	SetServiceCommonOptions(&opt.CommonOptions, oc, config.ServiceCommonOptions)
	opt.AutoSyncTable = true
	opt.SslCertfile = path.Join(constants.CertDir, constants.ServiceCertName)
	opt.SslKeyfile = path.Join(constants.CertDir, constants.ServiceKeyName)
	opt.Port = constants.MonitorPort
	return m.newServiceConfigMap(v1alpha1.MonitorComponentType, oc, opt), nil
}

func (m *monitorManager) getService(oc *v1alpha1.OnecloudCluster) []*corev1.Service {
	return []*corev1.Service{m.newSingleNodePortService(v1alpha1.MonitorComponentType, oc, constants.MonitorPort)}
}

func (m *monitorManager) getDeployment(oc *v1alpha1.OnecloudCluster, cfg *v1alpha1.OnecloudClusterConfig) (*apps.Deployment, error) {
	cf := func(volMounts []corev1.VolumeMount) []corev1.Container {
		return []corev1.Container{
			{
				Name:            "monitor",
				Image:           oc.Spec.Monitor.Image,
				ImagePullPolicy: oc.Spec.Monitor.ImagePullPolicy,
				Command:         []string{"/opt/yunion/bin/monitor", "--config", "/etc/yunion/monitor.conf"},
				VolumeMounts:    volMounts,
			},
		}
	}
	return m.newDefaultDeploymentNoInit(
		v1alpha1.MonitorComponentType, oc,
		NewVolumeHelper(oc, controller.ComponentConfigMapName(oc, v1alpha1.MonitorComponentType), v1alpha1.MonitorComponentType),
		oc.Spec.Monitor, cf)
}

func (m *monitorManager) getDeploymentStatus(oc *v1alpha1.OnecloudCluster) *v1alpha1.DeploymentStatus {
	return &oc.Status.Monitor
}
