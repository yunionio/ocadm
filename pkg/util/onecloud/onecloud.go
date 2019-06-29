package onecloud

import (
	"net/http"
	"strings"
	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"

	"yunion.io/x/onecloud/pkg/util/httputils"
)

const (
	NotFoundMsg = "NotFoundError"
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

func IsResourceExists(s *mcclient.ClientSession, manager modules.Manager, name string) (jsonutils.JSONObject, bool, error) {
	obj, err := manager.Get(s, name, nil)
	if err == nil {
		return obj, true, nil
	}
	if IsNotFoundError(err) {
		return nil, false, nil
	}
	return nil, false, err
}

func IsRoleExists(s *mcclient.ClientSession, roleName string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &modules.RolesV3, roleName)
}

func CreateRole(s *mcclient.ClientSession, roleName, description string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(roleName), "name")
	if description != "" {
		params.Add(jsonutils.NewString(description), "description")
	}
	return modules.RolesV3.Create(s, params)
}

func EnsureRole(s *mcclient.ClientSession, roleName, description string) (jsonutils.JSONObject, error) {
	obj, exists, err := IsRoleExists(s, roleName)
	if err != nil {
		return nil, err
	}
	if exists {
		return obj, nil
	}
	return CreateRole(s, roleName, description)
}

func IsServiceExists(s *mcclient.ClientSession, svcName string) (jsonutils.JSONObject, bool, error) {
	return IsResourceExists(s, &modules.ServicesV3, svcName)
}

func EnsureService(s *mcclient.ClientSession, svcName, svcType string) (jsonutils.JSONObject, error) {
	obj, exists, err := IsServiceExists(s, svcName)
	if err != nil {
		return nil, err
	}
	if exists {
		return obj, nil
	}
	return CreateService(s, svcName, svcType)
}

func CreateService(s *mcclient.ClientSession, svcName, svcType string) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(svcType), "type")
	params.Add(jsonutils.NewString(svcName), "name")
	params.Add(jsonutils.JSONTrue, "enabled")
	return modules.ServicesV3.Create(s, params)
}

func IsEndpointExists(s *mcclient.ClientSession, svcId, regionId, interfaceType string) (jsonutils.JSONObject, bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(svcId), "service_id")
	params.Add(jsonutils.NewString(regionId), "region_id")
	params.Add(jsonutils.NewString(interfaceType), "interface")
	eps, err := modules.EndpointsV3.List(s, params)
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
		return modules.EndpointsV3.Create(s, createParams)
	}
	epId, err := ep.GetString("id")
	if err != nil {
		return nil, err
	}
	updateParams := jsonutils.NewDict()
	updateParams.Add(jsonutils.NewString(url), "url")
	updateParams.Add(jsonutils.JSONTrue, "enabled")
	return modules.EndpointsV3.Update(s, epId, updateParams)
}
