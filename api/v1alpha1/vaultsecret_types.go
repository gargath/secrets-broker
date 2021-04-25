/*


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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SyncState string

const (
	// The in-cluster Secret is in sync with the source
	InSyncState SyncState = "InSync"

	// The in-cluster Secret is waiting to be updated
	SyncingState SyncState = "Synchronizing"

	// There was an error reconciling the Secret
	ErrorState SyncState = "Error"

	// The VaultSecret is pending reconciliation
	PendingState SyncState = "Pending"

	// The in-cluster Secret is stale, i.e. the source has changed but the Secret not been updated
	StaleState SyncState = "Stale"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VaultSecretSpec defines the desired state of VaultSecret
type VaultSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// VaultURL is the URL of the vault server that holds this secret. The controller must
	// hold configuration on how to authenticate to this vault in order to retrieve the secret
	VaultURL string `json:"vaultURL"`

	// SecretPath is the path within the specified Vault where the secrest value is held
	SecretPath string `json:"secretPath"`

	// Spec defines the structure of the managed Secret
	Spec SecretSpec `json:"spec"`
}

// SecretSpec defines the structure of the managed Secret
type SecretSpec struct {
	// Type is the type of the managed secret
	Type v1.SecretType `json:"type"`

	// FieldsRefs holds the field mappings from Vault to the managed Secret
	FieldRefs map[string]string `json:"fieldRefs"`
}

// VaultSecretStatus defines the observed state of VaultSecret
type VaultSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Phase SyncState `json:"phase"`

	// Conditions is the list of error conditions for this resource
	Conditions []*metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// VaultSecret is the Schema for the vaultsecrets API
type VaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultSecretSpec   `json:"spec,omitempty"`
	Status VaultSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VaultSecretList contains a list of VaultSecret
type VaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultSecret{}, &VaultSecretList{})
}
