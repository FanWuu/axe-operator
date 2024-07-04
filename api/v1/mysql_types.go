/*
Copyright 2024.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MysqlConf map[string]string

type RouterConf map[string]string

type Policy struct {
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default:="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	Labels            map[string]string   `json:"labels,omitempty"`
	Annotations       map[string]string   `json:"annotations,omitempty"`
	Affinity          *corev1.Affinity    `json:"affinity,omitempty"`
	PriorityClassName string              `json:"priorityClassName,omitempty"`
	Tolerations       []corev1.Toleration `json:"tolerations,omitempty"`

	// ExtraResources defines quotas for containers other than mysql or xenon.
	// These containers take up less resources, so quotas are set uniformly.
	// +optional
	// +kubebuilder:default:={requests: {cpu: "1", memory: "1Gi"}}
	ExtraResources corev1.ResourceRequirements `json:"extraResources,omitempty"`
}

// Persistence is the desired spec for storing mysql data. Only one of its
// members may be specified.
type Persistence struct {
	// Create a volume to store data.
	// +optional
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	// +kubebuilder:default:={"ReadWriteOnce"}
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// Size of persistent volume claim.
	// +optional
	// +kubebuilder:default:="10Gi"
	Size string `json:"size,omitempty"`
}

// MysqlSpec defines the desired state of Mysql
type MysqlSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of pods.
	// +optional
	// +kubebuilder:validation:Enum=0;3
	// +kubebuilder:default:=3
	Replica int32 `json:"replica,omitempty"`

	// +optional
	Router RouterOpts `json:"router,omitempty"`

	// +optional
	Mysql MysqlOpts `json:"mysql,omitempty"`

	// +optional
	// +kubebuilder:default:={imagePullPolicy: "IfNotPresent", extraResources: {requests: {cpu: "10m", memory: "32Mi"}}}
	PodPolicy Policy `json:"podpolicy,omitempty"`

	// +optional
	Persistence Persistence `json:"persistence,omitempty"`
}

type MysqlOpts struct {

	// The mysql image.
	// +optional
	// +kubebuilder:default:="mysql:8.0.32"
	MysqlImage string `json:"mysqlimage,omitempty"`

	// MysqlConfTemplate is the configmap name of the template for mysql config.
	// The configmap should contain the keys `mysql.cnf` and `plugin.cnf` at least, key `init.sql` is optional.
	// If empty, operator will generate a default template named <spec.metadata.name>-mysql.
	// +optional
	MysqlConfTemplate string `json:"mysqlConfTemplate,omitempty"`

	// If empty, operator will generate a default template named <spec.metadata.name>-mysql.
	// +optional
	PluginConfTemplate string `json:"pluginConfTemplate,omitempty"`
	// A map[string]string that will be passed to my.cnf file.
	// The key/value pairs is persisted in the configmap.
	// Delete key is not valid, it is recommended to edit the configmap directly.
	// +optional
	MysqlConf MysqlConf `json:"mysqlConf,omitempty"`

	// A map[string]string that will be passed to plugin.cnf file.
	// The key/value pairs is persisted in the configmap.
	// Delete key is not valid, it is recommended to edit the configmap directly.
	// +optional
	PluginConf MysqlConf `json:"pluginConf,omitempty"`

	// +optional
	// +kubebuilder:default:="axe_operator"
	RootPassword string `json:"rootPassword"`

	// +optional
	// +kubebuilder:default:="axe"
	MysqlUser string `json:"mysqlUser"`

	// +optional
	// +kubebuilder:default:="123456"
	MysqlPassword string `json:"mysqlPassword"`

	// The compute resource requirements.
	// +optional
	// +kubebuilder:default:={limits: {cpu: "2048m", memory: "2Gi"}, requests: {cpu: "1024m", memory: "256Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type RouterOpts struct {

	// The mysql-router image.
	// +optional
	// +kubebuilder:default:="mysql/mysql-router:latest"
	RouterImage string `json:"routerimage,omitempty"`

	// +optional
	// +kubebuilder:validation:Enum=0;1;2;3
	// +kubebuilder:default:=1
	Replica int32 `json:"replica,omitempty"`

	// The compute resource requirements.
	// +optional
	// +kubebuilder:default:={limits: {cpu: "2048m", memory: "2Gi"}, requests: {cpu: "1024m", memory: "256Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// MysqlConfTemplate is the configmap name of the template for mysql config.
	// The configmap should contain the keys `mysql.cnf` and `plugin.cnf` at least, key `init.sql` is optional.
	// If empty, operator will generate a default template named <spec.metadata.name>-mysql.
	// +optional
	RouterConfTemplate string `json:"mysqlConfTemplate,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// The key/value pairs is persisted in the configmap.
	// Delete key is not valid, it is recommended to edit the configmap directly.
	// +optional
	RouterConf RouterConf `json:"mysqlConf,omitempty"`
}

const (
	// ClusterInitState indicates whether the cluster is initializing.
	ClusterInitState string = "Initializing"
	// ClusterUpdateState indicates whether the cluster is being updated.
	ClusterUpdateState string = "Updating"
	// ClusterReadyState indicates whether all containers in the pod are ready.
	ClusterReadyState string = "Ready"
	// ClusterCloseState indicates whether the cluster is closed.
	ClusterCloseState string = "Closed"
	// ClusterScaleInState indicates whether the cluster replicas is decreasing.
	ClusterScaleInState string = "ScaleIn"
	// ClusterScaleOutState indicates whether the cluster replicas is increasing.
	ClusterScaleOutState string = "ScaleOut"
)

const (
	// ConditionInit indicates whether the cluster is initializing.
	ConditionInit string = "Initializing"
	// ConditionUpdate indicates whether the cluster is being updated.
	ConditionUpdate string = "Updating"
	// ConditionReady indicates whether all containers in the pod are ready.
	ConditionReady string = "Ready"
	// ConditionClose indicates whether the cluster is closed.
	ConditionClose string = "Closed"
	// ConditionError indicates whether there is an error in the cluster.
	ConditionError string = "Error"
	// ConditionScaleIn indicates whether the cluster replicas is decreasing.
	ConditionScaleIn string = "ScaleIn"
	// ConditionScaleOut indicates whether the cluster replicas is increasing.
	ConditionScaleOut string = "ScaleOut"
)

const (
	MgrNOTinstalled string = "MGR_NOT_INSTALLED"
	Mgrinstalled    string = "MGR_INSTALLED"
	MgrISinstall    string = "MGR_IS_INSTALL"
	MYSQLAPP        string = "mysql"
	MYSQLROUTERAPP  string = "mysql-router"
)

const (
	// NodeConditionLagged represents if the node is lagged.
	NodeConditionLagged string = "Lagged"
	// NodeConditionLeader represents if the node is leader or not.
	NodeConditionLeader string = "Leader"
	// NodeConditionReadOnly repesents if the node is read only or not
	NodeConditionReadOnly string = "ReadOnly"
	// NodeConditionReplicating represents if the node is replicating or not.
	NodeConditionReplicating string = "Replicating"
)

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ReadyNodes represents number of the nodes that are in ready state.
	ReadyNodes int `json:"readyNodes,omitempty"`
	// State
	State string `json:"state,omitempty"`
	// Conditions contains the list of the cluster conditions fulfilled.
	// Nodes contains the list of the node status fulfilled.
	Nodes string `json:"nodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.readyNodes
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The cluster status"
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Current",type="integer",JSONPath=".status.readyNodes",description="The number of current replicas"
// +kubebuilder:printcolumn:name="Leader",type="string",JSONPath=".status.nodes",description="Name of the leader node"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=mysql
// Mysql is the Schema for the mysqls API
type Mysql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlSpec   `json:"spec,omitempty"`
	Status MysqlStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlList contains a list of Mysql
type MysqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysql `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mysql{}, &MysqlList{})
}
