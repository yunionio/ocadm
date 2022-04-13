package onecloud

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	nodeutil "k8s.io/kubernetes/pkg/util/node"

	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud-operator/pkg/apis/constants"
	"yunion.io/x/onecloud-operator/pkg/manager/component"
	"yunion.io/x/onecloud/pkg/hostman/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modulebase"
	compute_modules "yunion.io/x/onecloud/pkg/mcclient/modules/compute"
	identity_modules "yunion.io/x/onecloud/pkg/mcclient/modules/identity"
	"yunion.io/x/onecloud/pkg/util/httputils"
)

const (
	NotFoundMsg  = "NotFoundError"
	HostConfFile = "/etc/yunion/host.conf"
)

func IsNotFoundError(err error) bool {
	if httpErr, ok := err.(*httputils.JSONClientError); ok {
		if httpErr.Code == http.StatusNotFound {
			return true
		}
	}
	if strings.Contains(err.Error(), NotFoundMsg) {
		return true
	}
	return false
}

func IsResourceExists(s *mcclient.ClientSession, manager modulebase.Manager, name string) (jsonutils.JSONObject, bool, error) {
	obj, err := manager.Get(s, name, nil)
	if err == nil {
		return obj, true, nil
	}
	if IsNotFoundError(err) {
		return nil, false, nil
	}
	return nil, false, err
}

func EnsureResource(
	s *mcclient.ClientSession,
	man modulebase.Manager,
	name string,
	createFunc func() (jsonutils.JSONObject, error),
) (jsonutils.JSONObject, error) {
	obj, exists, err := IsResourceExists(s, man, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return obj, nil
	}
	return createFunc()
}

func DeleteResource(
	s *mcclient.ClientSession,
	man modulebase.Manager,
	name string,
) error {
	obj, exists, err := IsResourceExists(s, man, name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	id, _ := obj.GetString("id")
	_, err = man.Delete(s, id, nil)
	return err
}

func IsRoleExists(s *mcclient.ClientSession, roleName string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &identity_modules.RolesV3, roleName)
}

func CreateRole(s *mcclient.ClientSession, roleName, description string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(roleName), "name")
	if description != "" {
		params.Add(jsonutils.NewString(description), "description")
	}
	return identity_modules.RolesV3.Create(s, params)
}

func EnsureRole(s *mcclient.ClientSession, roleName, description string) (jsonutils.JSONObject, error) {
	return EnsureResource(s, &identity_modules.RolesV3, roleName, func() (jsonutils.JSONObject, error) {
		return CreateRole(s, roleName, description)
	})
}

func IsServiceExists(s *mcclient.ClientSession, svcName string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &identity_modules.ServicesV3, svcName)
}

func EnsureService(s *mcclient.ClientSession, svcName, svcType string) (jsonutils.JSONObject, error) {
	return EnsureResource(s, &identity_modules.ServicesV3, svcName, func() (jsonutils.JSONObject, error) {
		return CreateService(s, svcName, svcType)
	})
}

func CreateService(s *mcclient.ClientSession, svcName, svcType string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(svcType), "type")
	params.Add(jsonutils.NewString(svcName), "name")
	params.Add(jsonutils.JSONTrue, "enabled")
	return identity_modules.ServicesV3.Create(s, params)
}

func IsEndpointExists(s *mcclient.ClientSession, svcId, regionId, interfaceType string) (jsonutils.JSONObject, bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(svcId), "service_id")
	params.Add(jsonutils.NewString(regionId), "region_id")
	params.Add(jsonutils.NewString(interfaceType), "interface")
	eps, err := identity_modules.EndpointsV3.List(s, params)
	if err != nil {
		return nil, false, err
	}
	if len(eps.Data) == 0 {
		return nil, false, nil
	}
	return eps.Data[0], true, nil
}

func EnsureEndpoint(s *mcclient.ClientSession, svcId, regionId, interfaceType, url string) (jsonutils.JSONObject, error) {
	ep, exists, err := IsEndpointExists(s, svcId, regionId, interfaceType)
	if err != nil {
		return nil, err
	}
	if !exists {
		createParams := jsonutils.NewDict()
		createParams.Add(jsonutils.NewString(svcId), "service_id")
		createParams.Add(jsonutils.NewString(regionId), "region_id")
		createParams.Add(jsonutils.NewString(interfaceType), "interface")
		createParams.Add(jsonutils.NewString(url), "url")
		createParams.Add(jsonutils.JSONTrue, "enabled")
		return identity_modules.EndpointsV3.Create(s, createParams)
	}
	epId, err := ep.GetString("id")
	if err != nil {
		return nil, err
	}
	updateParams := jsonutils.NewDict()
	updateParams.Add(jsonutils.NewString(url), "url")
	updateParams.Add(jsonutils.JSONTrue, "enabled")
	return identity_modules.EndpointsV3.Update(s, epId, updateParams)
}

func IsUserExists(s *mcclient.ClientSession, username string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &identity_modules.UsersV3, username)
}

func CreateUser(s *mcclient.ClientSession, username string, password string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(username), "name")
	params.Add(jsonutils.NewString(password), "password")
	return identity_modules.UsersV3.Create(s, params)
}

func ChangeUserPassword(s *mcclient.ClientSession, username string, password string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(password), "password")
	return identity_modules.UsersV3.Update(s, username, params)
}

func ProjectAddUser(s *mcclient.ClientSession, projectId string, userId string, roleId string) error {
	_, err := identity_modules.RolesV3.PutInContexts(s, roleId, nil,
		[]modulebase.ManagerContext{
			{InstanceManager: &identity_modules.Projects, InstanceId: projectId},
			{InstanceManager: &identity_modules.UsersV3, InstanceId: userId},
		})
	return err
}

func IsZoneExists(s *mcclient.ClientSession, zone string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &compute_modules.Zones, zone)
}

func CreateZone(s *mcclient.ClientSession, zone string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(zone), "name")
	return compute_modules.Zones.Create(s, params)
}

func IsWireExists(s *mcclient.ClientSession, wire string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &compute_modules.Wires, wire)
}

func CreateWire(s *mcclient.ClientSession, zone string, wire string, bw int, vpc string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(wire), "name")
	params.Add(jsonutils.NewInt(int64(bw)), "bandwidth")
	params.Add(jsonutils.NewString(vpc), "vpc")
	return compute_modules.Wires.CreateInContext(s, params, &compute_modules.Zones, zone)
}

func IsNetworkExists(s *mcclient.ClientSession, net string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &compute_modules.Networks, net)
}

func CreateNetwork(
	s *mcclient.ClientSession,
	name string,
	gateway string,
	serverType string,
	wireId string,
	maskLen int,
	startIp string,
	endIp string,
) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(name), "name")
	params.Add(jsonutils.NewString(startIp), "guest_ip_start")
	params.Add(jsonutils.NewString(endIp), "guest_ip_end")
	params.Add(jsonutils.NewInt(int64(maskLen)), "guest_ip_mask")
	if gateway != "" {
		params.Add(jsonutils.NewString(gateway), "guest_gateway")
	}
	if serverType != "" {
		params.Add(jsonutils.NewString(serverType), "server_type")
	}
	return compute_modules.Networks.CreateInContext(s, params, &compute_modules.Wires, wireId)
}

func NetworkPrivate(s *mcclient.ClientSession, name string) (jsonutils.JSONObject, error) {
	return compute_modules.Networks.PerformAction(s, "private", name, nil)
}

func CreateRegion(s *mcclient.ClientSession, region, zone string) (jsonutils.JSONObject, error) {
	if zone != "" {
		region = mcclient.RegionID(region, zone)
	}
	obj, err := identity_modules.Regions.Get(s, region, nil)
	if err == nil {
		// region already exists
		return obj, nil
	}
	if !IsNotFoundError(err) {
		return nil, err
	}
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(region), "id")
	return identity_modules.Regions.Create(s, params)
}

func IsSchedtagExists(s *mcclient.ClientSession, name string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &compute_modules.Schedtags, name)
}

func CreateSchedtag(s *mcclient.ClientSession, name string, strategy string, description string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(name), "name")
	params.Add(jsonutils.NewString(strategy), "default_strategy")
	params.Add(jsonutils.NewString(description), "description")
	return compute_modules.Schedtags.Create(s, params)
}

func EnsureSchedtag(s *mcclient.ClientSession, name string, strategy string, description string) (jsonutils.JSONObject, error) {
	return EnsureResource(s, &compute_modules.Schedtags, name, func() (jsonutils.JSONObject, error) {
		return CreateSchedtag(s, name, strategy, description)
	})
}

func IsDynamicSchedtagExists(s *mcclient.ClientSession, name string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &compute_modules.Dynamicschedtags, name)
}

func CreateDynamicSchedtag(s *mcclient.ClientSession, name, schedtag, condition string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(name), "name")
	params.Add(jsonutils.NewString(schedtag), "schedtag")
	params.Add(jsonutils.NewString(condition), "condition")
	params.Add(jsonutils.JSONTrue, "enabled")
	return compute_modules.Dynamicschedtags.Create(s, params)
}

func EnsureDynamicSchedtag(s *mcclient.ClientSession, name, schedtag, condition string) (jsonutils.JSONObject, error) {
	return EnsureResource(s, &compute_modules.Dynamicschedtags, name, func() (jsonutils.JSONObject, error) {
		return CreateDynamicSchedtag(s, name, schedtag, condition)
	})
}

func GetEndpointsByService(s *mcclient.ClientSession, serviceName string) ([]jsonutils.JSONObject, error) {
	obj, err := identity_modules.ServicesV3.Get(s, serviceName, nil)
	if err != nil {
		return nil, err
	}
	svcId, _ := obj.GetString("id")
	searchParams := jsonutils.NewDict()
	searchParams.Add(jsonutils.NewString(svcId), "service_id")
	ret, err := identity_modules.EndpointsV3.List(s, searchParams)
	if err != nil {
		return nil, err
	}
	return ret.Data, nil
}

func DisableService(s *mcclient.ClientSession, id string) error {
	params := jsonutils.NewDict()
	params.Add(jsonutils.JSONFalse, "enabled")
	_, err := identity_modules.ServicesV3.Patch(s, id, params)
	return err
}

func DisableEndpoint(s *mcclient.ClientSession, id string) error {
	params := jsonutils.NewDict()
	params.Add(jsonutils.JSONFalse, "enabled")
	_, err := identity_modules.EndpointsV3.Patch(s, id, params)
	return err
}

func DeleteServiceEndpoints(s *mcclient.ClientSession, serviceName string) error {
	endpoints, err := GetEndpointsByService(s, serviceName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}
	for _, ep := range endpoints {
		id, _ := ep.GetString("id")
		tmpId := id
		if err := DisableEndpoint(s, tmpId); err != nil {
			return err
		}
		if _, err := identity_modules.EndpointsV3.Delete(s, id, nil); err != nil {
			return err
		}
	}
	if err := DisableService(s, serviceName); err != nil {
		return err
	}
	return DeleteResource(s, &identity_modules.ServicesV3, serviceName)
}

type HostCfg struct {
	EnableHost bool

	LocalImagePath []string
	Networks       []string
	Hostname       string
}

func GenerateDefaultHostConfig(cfg *HostCfg) error {
	var o = new(options.SHostOptions)
	component.SetOptionsDefault(o, "")
	o.LocalImagePath = cfg.LocalImagePath
	o.Networks = cfg.Networks
	if len(cfg.Hostname) > 0 {
		o.Hostname = cfg.Hostname
	} else {
		hostname, err := nodeutil.GetHostname("")
		if err != nil {
			return errors.Wrap(err, "get hostname")
		} else if strings.Contains(hostname, ".") {
			hostname = strings.Split(hostname, ".")[0]
			o.Hostname = hostname
		}
	}

	o.ReportInterval = 60
	o.BridgeDriver = "openvswitch"
	o.ServersPath = "/opt/cloud/workspace/servers"
	o.OvmfPath = "/opt/cloud/contrib/OVMF.fd"
	if len(o.LocalImagePath) == 0 {
		o.LocalImagePath = []string{"/opt/cloud/workspace/disks"}
	}
	o.ImageCachePath = "/opt/cloud/workspace/disks/image_cache"
	o.AgentTempPath = "/opt/cloud/workspace/disks/agent_tmp"
	o.Rack = "rack0"
	o.Slots = "slot0"
	o.LinuxDefaultRootUser = true
	o.EnableOpenflowController = false
	o.BlockIoScheduler = "cfq"
	o.EnableTemplateBacking = true
	o.DefaultQemuVersion = "4.2.0"
	o.EnableRemoteExecutor = true
	o.OvnSouthDatabase = fmt.Sprintf("tcp:default-ovn-north:%d", constants.OvnSouthDbPort)
	if err := os.MkdirAll("/opt/cloud", os.ModePerm); err != nil {
		return errors.Wrap(err, "mkdir /opt/cloud")
	}
	if err := os.MkdirAll("/etc/yunion", os.ModePerm); err != nil {
		return errors.Wrap(err, "mkdir /etc/yunion")
	}
	if _, err := os.Stat(HostConfFile); !os.IsNotExist(err) {
		os.Rename(HostConfFile, HostConfFile+".backup")
	}
	yamlStr := jsonutils.Marshal(o).YAMLString()
	return ioutil.WriteFile(HostConfFile, []byte(yamlStr), 0664)
}
