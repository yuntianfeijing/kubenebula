/*

 Copyright 2019 The KubeSphere Authors.

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
package constants

const (
	APIVersion    = "v1alpha1"
	ClusterAdmin  = "cluster-admin"
	TeamAdmin     = "team-admin"
	TeamRegular   = "team-regular"
	TeamViewer    = "team-viewer"
	AdminUserName = "admin"

	TeamLabelKey             = "kubenebula.io/teambase64"  //Team Label in namespace
	DisplayNameAnnotationKey = "kubenebula.io/alias-name"  //别名
	DescriptionAnnotationKey = "kubenebula.io/description" //描述
	CreatorAnnotationKey     = "kubenebula.io/creator"     //创建者
	AvatarAnnotationKey      = "kubenebula.io/avatar"
	TeamAnnotationKey        = "kubenebula.io/team" //Team Label in namespace
	System                   = "system"             //默认的系统创建者，创建的资源视为不可被用户删除的资源

	ResourceLabel              = "kubenebula.io/resource"
	ResourceClusterRole        = "clusterrole"
	ResourceRole               = "role"
	ResourceClusterRoleBinding = "clusterrolebinding"
	ResourceRoleBinding        = "rolebinding"
	UserNameHeader             = "X-Token-Username"
	TenantResourcesTag         = "Tenant Resources"
	NamespaceResourcesTag      = "Namespace Resources"
	ClusterResourcesTag        = "Cluster Resources"
	ComponentStatusTag         = "Component Status"
	VerificationTag            = "Verification"
	ClusterMetricsTag          = "Cluster Metrics"
	NodeMetricsTag             = "Node Metrics"
	NamespaceMetricsTag        = "Namespace Metrics"
	PodMetricsTag              = "Pod Metrics"
	PVCMetricsTag              = "PVC Metrics"
	ContainerMetricsTag        = "Container Metrics"
	WorkloadMetricsTag         = "Workload Metrics"
	TeamMetricsTag             = "Team Metrics"
	ComponentMetricsTag        = "Component Metrics"
	LogQueryTag                = "Log Query"
	FluentBitSetting           = "Fluent Bit Setting"
)

var (
	TeamRoles = []string{TeamAdmin, TeamRegular, TeamViewer}
	//SystemNamespaces = []string{KubeSphereNamespace, KubeSphereLoggingNamespace, KubeSphereMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, KubesphereDevOpsNamespace, PorterNamespace}
)
