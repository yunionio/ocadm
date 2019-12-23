package cmd

import (
	operatorconstants "yunion.io/x/onecloud-operator/pkg/apis/constants"
)

type hostEnableData struct {
	nodesBaseData
}

func (h *hostEnableData) Command() string {
	return "enable-host-agent"
}

func (h *hostEnableData) Short() string {
	return "Run this command to select node enable host agent"
}

func (h *hostEnableData) GetLabels() map[string]string {
	return map[string]string{
		operatorconstants.OnecloudEnableHostLabelKey: "enable",
	}
}

type hostDisableData struct {
	nodesBaseData
}

func (h *hostDisableData) Command() string {
	return "disable-host-agent"
}

func (h *hostDisableData) Short() string {
	return "Run this command to select node disable host agent"
}

func (h *hostDisableData) GetLabels() map[string]string {
	return map[string]string{
		operatorconstants.OnecloudEnableHostLabelKey: "disable",
	}
}

type onecloudControllerEnableData struct {
	nodesBaseData
}

func (h *onecloudControllerEnableData) Command() string {
	return "enable-onecloud-controller"
}

func (h *onecloudControllerEnableData) Short() string {
	return "Run this command to select node enable onecloud controller"
}

func (h *onecloudControllerEnableData) GetLabels() map[string]string {
	return map[string]string{
		operatorconstants.OnecloudControllerLabelKey: "enable",
	}
}

type onecloudControllerDisableData struct {
	nodesBaseData
}

func (h *onecloudControllerDisableData) Command() string {
	return "disable-onecloud-controller"
}

func (h *onecloudControllerDisableData) Short() string {
	return "Run this command to select node disable onecloud controller"
}

func (h *onecloudControllerDisableData) GetLabels() map[string]string {
	return map[string]string{
		operatorconstants.OnecloudControllerLabelKey: "disable",
	}
}
