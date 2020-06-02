/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package keepalived

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

type Member struct {
	Name    string
	PeerURL string
}

func CreateLocalKeepalivedStaticPodManifestFile(manifestDir string, nodeName string, cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint) error {
	// if cfg.Keepalived.External != nil {
	// 	return errors.New("keepalived static pod manifest cannot be generated for cluster using external keepalived")
	// }
	// gets keepalived StaticPodSpec
	// _ = GetKeepalivedPodSpec(cfg, endpoint, nodeName, []Member{})

	// writes keepalived StaticPod to disk
	// TODO 1st args should be const.
	// if err := staticpodutil.WriteStaticPodToDisk("keepalived", manifestDir, spec); err != nil {
	// 	return err
	// }

	// klog.V(1).Infof("[keepalived] wrote Static Pod manifest for a local keepalived member to %q\n", kubeadmconstants.GetStaticPodFilepath(kubeadmconstants.Keepalived, manifestDir))
	return nil
}

// GetKeepalivedPodSpec returns the keepalived static Pod actualized to the context of the current configuration
// NB. GetKeepalivedPodSpec methods holds the information about how kubeadm creates keepalived static pod manifests.
//func GetKeepalivedPodSpec(cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint, nodeName string, initialCluster []Member) error {
//
//	keepalivedMounts := map[string]v1.Volume{}
//	return staticpodutil.ComponentPod(v1.Container{
//		Name:            "keepalived",
//		Command:         getKeepalivedCommand(cfg, endpoint, nodeName, initialCluster),
//		Image:           images.GetKeepalivedImage(cfg),
//		ImagePullPolicy: v1.PullIfNotPresent,
//		// Mount the keepalived datadir path read-write so keepalived can store data in a more persistent manner
//		VolumeMounts: []v1.VolumeMount{
//			staticpodutil.NewVolumeMount(keepalivedVolumeName, cfg.Keepalived.Local.DataDir, false),
//			staticpodutil.NewVolumeMount(certsVolumeName, cfg.CertificatesDir+"/keepalived", false),
//		},
//		LivenessProbe: staticpodutil.KeepalivedProbe(
//			&cfg.Keepalived, kubeadmconstants.KeepalivedListenClientPort, cfg.CertificatesDir,
//			kubeadmconstants.KeepalivedCACertName, kubeadmconstants.KeepalivedHealthcheckClientCertName, kubeadmconstants.KeepalivedHealthcheckClientKeyName,
//		),
//	}, keepalivedMounts)
//	return nil
//}
