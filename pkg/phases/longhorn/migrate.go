package longhorn

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"yunion.io/x/onecloud-operator/pkg/apis/onecloud/v1alpha1"

	"yunion.io/x/log"
	"yunion.io/x/ocadm/pkg/apis/constants"
	"yunion.io/x/ocadm/pkg/util/kubectl"
)

func MigrateLonghornDataPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "migrate-longhorn-data",
		Short: "migrate lognhorn data",
		Phases: []workflow.Phase{
			{
				Name: "longhorn-create-pvc",
				Run:  runLonghornEnsurePvc,
			},
			{
				Name: "migrate-longhorn-data",
				Run:  runLonghornMigrateData,
			},
		},
	}
}

type MigrateToLonghornConfig interface {
	KubectlClient() (*kubectl.Client, error)
	SourcePVC() string
	ClientSet() (*clientset.Clientset, error)
	GetImageRepository() string
	DeleteMigartePodInTheEnd() bool
	MigrateToSourcePvc() bool
}

var destPVCName string

func isPVCMounted(desc string) bool {
	lines := strings.Split(desc, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Mounted By:") {
			segs := strings.Split(lines[i], " ")
			if strings.TrimSpace(segs[len(segs)-1]) != "<none>" {
				return true
			}
		}
	}
	return false
}

func runLonghornEnsurePvc(c workflow.RunData) error {
	data, ok := c.(MigrateToLonghornConfig)
	if !ok {
		return errors.New("addon phase invoked with an invalid data struct")
	}

	migrateToSourcePvc := data.MigrateToSourcePvc()
	cli, err := data.ClientSet()
	if err != nil {
		return err
	}
	storageClass := constants.LonghornStorageClass
	srcPVC := data.SourcePVC()
	sp, err := cli.CoreV1().PersistentVolumeClaims(constants.OnecloudNamespace).Get(srcPVC, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get source pvc %s failed", srcPVC)
	}
	if *sp.Spec.StorageClassName == v1alpha1.DefaultStorageClass {
		destPVCName = fmt.Sprintf("%s-%s", srcPVC, storageClass)
	} else {
		segs := strings.Split(srcPVC, "-")
		srcName := segs[:len(segs)-1]
		destPVCName = fmt.Sprintf("%s-%s", srcName, storageClass)
	}
	if migrateToSourcePvc {
		_, err := cli.CoreV1().PersistentVolumeClaims(constants.OnecloudNamespace).Get(destPVCName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "migrate to source pvc get longhorn pvc %s failed", destPVCName)
		}
		return nil
	}

	kubecli, err := data.KubectlClient()
	if err != nil {
		return errors.Wrap(err, "get kubectl client")
	}
	out, err := kubecli.Describe("pvc", srcPVC, constants.OnecloudNamespace)
	if err != nil {
		return errors.Wrap(err, "descibe pvc")
	}
	if isPVCMounted(string(out)) {
		return errors.Errorf("pvc %s is mounted, can't do migrate", srcPVC)
	}

	destPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      destPVCName,
			Namespace: constants.OnecloudNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: sp.Spec.Resources.Requests[corev1.ResourceStorage],
				},
			},
			StorageClassName: &storageClass,
		},
	}
	destPVC, err = cli.CoreV1().PersistentVolumeClaims(constants.OnecloudNamespace).Create(destPVC)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "create target pvc %s failed", destPVCName)
	}
	klog.Infof("PVC %s created", destPVCName)
	return nil
}

const (
	SYNC_IMAGE = "rsync-ssh"
	POD_NAME   = "migrate-pv-data"
)

func getSyncImage(imageRepository string) string {
	return fmt.Sprintf("%s:latest", path.Join(imageRepository, SYNC_IMAGE))
}

func runLonghornMigrateData(c workflow.RunData) error {
	data, ok := c.(MigrateToLonghornConfig)
	if !ok {
		return errors.New("addon phase invoked with an invalid data struct")
	}
	cli, err := data.ClientSet()
	if err != nil {
		return err
	}
	kubecli, err := data.KubectlClient()
	if err != nil {
		return errors.Wrap(err, "get kubectl client")
	}
	srcPVC := data.SourcePVC()
	migrateDestPVC := destPVCName
	if data.MigrateToSourcePvc() {
		srcPVC, migrateDestPVC = migrateDestPVC, srcPVC
	}

	migratePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      POD_NAME,
			Namespace: constants.OnecloudNamespace,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "src",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: srcPVC,
							ReadOnly:  false,
						},
					},
				},
				{
					Name: "dest",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: migrateDestPVC,
							ReadOnly:  false,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "sync-data",
					Image:   getSyncImage(data.GetImageRepository()),
					Command: []string{"sh", "-c", "tail -f /dev/null"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "src",
							MountPath: "/src",
						},
						{
							Name:      "dest",
							MountPath: "/dest",
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
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
		},
	}
	_, err = cli.CoreV1().Pods(constants.OnecloudNamespace).Create(migratePod)
	if err != nil {
		return err
	}
	if data.DeleteMigartePodInTheEnd() {
		defer func() {
			// delete migrate pod
			e := cli.CoreV1().Pods(constants.OnecloudNamespace).Delete(POD_NAME, &metav1.DeleteOptions{})
			if e != nil {
				log.Errorf("Delete migrate pod failed %s", err)
			}
		}()
	}

	var ch = make(chan error)
	sync := func() {
		cmd := kubecli.Exec(POD_NAME, "", constants.OnecloudNamespace,
			[]string{"rsync", "--info=progress2", "--info=name0", "-av", "/src/", "/dest/"})
		err := cmd.Start()
		if err != nil {
			log.Errorln(err)
			ch <- err
			return
		}
		if err := cmd.Wait(); err != nil {
			log.Errorln(err)
			ch <- err
			return
		}
		log.Infoln("rsync done")
		ch <- nil
	}

	err = wait.PollImmediate(3*time.Second, time.Hour*24, func() (done bool, err error) {
		pod, err := cli.CoreV1().Pods(constants.OnecloudNamespace).Get(POD_NAME, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			log.Infoln("start sync data")
			go sync()
			return true, nil
		case corev1.PodPending:
			log.Infof("pod %s status pending", POD_NAME)
		case corev1.PodFailed, corev1.PodUnknown:
			return false, errors.Errorf("pod status %s, failed to sync data", pod.Status.Phase)
		}
		return false, nil
	})
	if err != nil {
		return err
	} else {
		err = <-ch
		if err != nil {
			return err
		}
	}

	log.Infof("Pvc migrate successed, new pvc name is %s", destPVCName)
	return nil
}
