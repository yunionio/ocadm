package keystone

import (
	"golang.org/x/sync/errgroup"

	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

const (
	PolicyDomainAdmin = `
# project owner, allow do any with her project resources
roles:
  - domainadmin
scope: domain
policy:
  '*': allow
`
	PolicyMember = `
# rbac for normal user, not allow for delete
scope: project
policy:
  '*':
    '*':
      '*': allow
      delete: deny
  compute:
    networks:
      create: deny
      perform: deny
      delete: deny
      update: deny
`
	PolicyProjectFA = `
# project finance administrator, allow any operation in meter
roles:
  - fa
scope: project
policy:
  meter: allow
`
	PolicyProjectOwner = `
# project owner, allow do any with her project resources
roles:
  - project_owner
  - admin
scope: project
policy:
  '*': allow
  compute:
    networks:
      create: deny
      perform: deny
      delete: deny
      update: deny
`
	PolicyProjectSA = `
# project system administrator, allow any operation in compute, image, k8s
roles:
  - sa
scope: project
policy:
  compute: allow
  image: allow
  k8s: owner
  compute:
    networks:
      create: deny
      perform: deny
      delete: deny
      update: deny
`
	PolicySysAdmin = `
# system wide administrator, root of the platform, can do anything
projects:
  - system
roles:
  - admin
scope: system
policy:
  '*': allow
`
	PolicySysFA = `
# system wide financial administrator, can do anything wrt billing&metering
projects:
  - system
roles:
  - fa
scope: system
policy:
  meter: allow
`
	PolicySysSA = `
# system wide ops administrator, can do anything wrt compute/image/k8s
projects:
  - system
roles:
  - sa
scope: system
policy:
  compute: allow
  image: allow
  k8s: allow
`
)

const (
	RoleAdmin        = "admin"
	RoleFA           = "fa"
	RoleSA           = "sa"
	RoleProjectOwner = "project_owner"
	RoleMember       = "member"
	RoleDomainAdmin  = "domainadmin"

	PolicyTypeDomainAdmin  = "domainadmin"
	PolicyTypeMember       = "member"
	PolicyTypeProjectFA    = "projectfa"
	PolicyTypeProjectOwner = "projectowner"
	PolicyTypeProjectSA    = "projectsa"
	PolicyTypeSysAdmin     = "sysadmin"
	PolicyTypeSysFA        = "sysfa"
	PolicyTypeSysSA        = "syssa"
)

var (
	PublicRoles = []string{
		RoleFA,
		RoleSA,
		RoleProjectOwner,
		RoleMember,
		RoleDomainAdmin,
	}
	PublicPolicies = []string{
		PolicyTypeDomainAdmin, PolicyTypeProjectOwner,
		PolicyTypeProjectSA, PolicyTypeProjectFA,
		PolicyTypeMember,
	}
)

type Policies map[string]string

var DefaultPolicies Policies
var DefaultRoles map[string]string

func init() {
	DefaultPolicies = map[string]string{
		PolicyTypeDomainAdmin:  PolicyDomainAdmin,
		PolicyTypeMember:       PolicyMember,
		PolicyTypeProjectFA:    PolicyProjectFA,
		PolicyTypeProjectOwner: PolicyProjectOwner,
		PolicyTypeProjectSA:    PolicyProjectSA,
		PolicyTypeSysAdmin:     PolicySysAdmin,
		PolicyTypeSysFA:        PolicySysFA,
		PolicyTypeSysSA:        PolicySysSA,
	}

	DefaultRoles = map[string]string{
		RoleAdmin:        "系统管理员",
		RoleFA:           "财务管理员",
		RoleSA:           "运维管理员",
		RoleProjectOwner: "项目主管",
		RoleMember:       "普通成员",
		RoleDomainAdmin:  "域管理员",
	}
}

func PolicyCreate(s *mcclient.ClientSession, policyType string, content string, enable bool) (jsonutils.JSONObject, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(policyType), "type")
	params.Add(jsonutils.NewString(content), "policy")
	if enable {
		params.Add(jsonutils.JSONTrue, "enabled")
	} else {
		params.Add(jsonutils.JSONFalse, "enabled")
	}
	return modules.Policies.Create(s, params)
}

func PoliciesPublic(s *mcclient.ClientSession, types []string) error {
	var errgrp errgroup.Group
	for _, t := range types {
		pt := t
		errgrp.Go(func() error {
			_, err := modules.Policies.PerformAction(s, pt, "public", nil)
			return err
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}
	return nil
}

func RolesPublic(s *mcclient.ClientSession, roles []string) error {
	var errgrp errgroup.Group
	for _, t := range roles {
		pt := t
		errgrp.Go(func() error {
			_, err := modules.RolesV3.PerformAction(s, pt, "public", nil)
			return err
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}
	return nil
}
