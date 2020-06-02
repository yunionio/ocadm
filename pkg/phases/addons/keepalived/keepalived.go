package keepalived

import (
	"fmt"
	"runtime"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"yunion.io/x/ocadm/pkg/phases/addons"
)

type KeepalivedConfig struct {
	Vip       string
	Interface string
}

func Trace(msg string) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fmt.Printf("[ rexxer %s ] %s:%d %s\n", msg, frame.File, frame.Line, frame.Function)
}

func NewKeepalivedConfig(cfg *kubeadmapi.ClusterConfiguration) addons.Configer {
	Trace("")
	config := &KeepalivedConfig{}
	return config
}

func (c KeepalivedConfig) Name() string {
	return "keepalived"
}

func (c KeepalivedConfig) GenerateYAML() (string, error) {
	Trace("")
	return addons.CompileTemplateFromMap(KeepalivedTemplate, c)
}
