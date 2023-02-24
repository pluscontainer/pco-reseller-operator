/*
Copyright © 2023 PlusServer GmbH

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

package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	Region      string `json:"region,omitempty"`
	Description string `json:"description,omitempty"`
	//Enabled     *bool            `json:"enabled,omitempty"`
	Quotas *QuotaCollection `json:"quotas,omitempty"`
}

type QuotaCollection struct {
	Compute *ComputeQuotas `json:"compute,omitempty"`
	Network *NetworkQuotas `json:"network,omitempty"`
	Volume  *VolumeQuotas  `json:"volume,omitempty"`
}

// ComputeQuotas defines model for ComputeQuotas.
type ComputeQuotas struct {
	// Number of cores between 0 and 500
	Cores    *int `json:"cores,omitempty"`
	FixedIps *int `json:"fixed_ips,omitempty"`

	// The number of allowed floating IP addresses for each project
	FloatingIps              *int `json:"floating_ips,omitempty"`
	InjectedFileContentBytes *int `json:"injected_file_content_bytes,omitempty"`
	InjectedFilePathBytes    *int `json:"injected_file_path_bytes,omitempty"`
	InjectedFiles            *int `json:"injected_files,omitempty"`
	Instances                int  `json:"instances"`
	KeyPairs                 int  `json:"key_pairs"`
	MetadataItems            int  `json:"metadata_items"`

	// Maximum amount of RAM in MiB
	Ram                *int `json:"ram,omitempty"`
	SecurityGroupRules *int `json:"security_group_rules,omitempty"`
	SecurityGroups     *int `json:"security_groups,omitempty"`
	ServerGroupMembers *int `json:"server_group_members,omitempty"`
	ServerGroups       int  `json:"server_groups"`
}

// VolumeQuotas defines model for VolumeQuotas.
type VolumeQuotas struct {
	BackupGigabytes int `json:"backup_gigabytes"`
	Backups         int `json:"backups"`

	// Maximum amount of available Storage
	Gigabytes          *int `json:"gigabytes,omitempty"`
	Groups             *int `json:"groups,omitempty"`
	PerVolumeGigabytes *int `json:"per_volume_gigabytes,omitempty"`

	// Maximum amount of snapshots
	Snapshots *int `json:"snapshots,omitempty"`
	Volumes   int  `json:"volumes"`
}

// NetworkQuotas defines model for NetworkQuotas.
type NetworkQuotas struct {
	// The number of floating IP addresses allowed for each project.A value of -1 means no limit
	Floatingip *int `json:"floatingip,omitempty"`
	Network    int  `json:"network"`
	Port       *int `json:"port,omitempty"`

	// The number of role-based access control (RBAC) policies for each project
	RbacPolicy        *int `json:"rbac_policy,omitempty"`
	Router            int  `json:"router"`
	SecurityGroup     int  `json:"security_group"`
	SecurityGroupRule int  `json:"security_group_rule"`
	Subnet            int  `json:"subnet"`
	Subnetpool        *int `json:"subnetpool,omitempty"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

func (r *Project) IsReady() bool {
	return meta.IsStatusConditionTrue(r.Status.Conditions, string(ProjectReady)) && meta.IsStatusConditionTrue(r.Status.Conditions, string(RegionReady))
}

func (r *Project) UpdateProjectCondition(ctx context.Context, reconcileClient client.Client, reason ProjectReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(ProjectReady), reason.projectStatus(), string(reason), message)
}

func (r *Project) UpdateRegionCondition(ctx context.Context, reconcileClient client.Client, reason RegionReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(RegionReady), reason.regionStatus(), string(reason), message)
}

func (r *Project) updateCondition(ctx context.Context, reconcileClient client.Client, typeString string, status v1.ConditionStatus, reason string, message string) error {
	oldProject := r.DeepCopy()

	meta.SetStatusCondition(&r.Status.Conditions, v1.Condition{
		Type:    typeString,
		Status:  status,
		Reason:  reason,
		Message: message,
	})

	return reconcileClient.Status().Patch(ctx, r, client.MergeFrom(oldProject))
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
