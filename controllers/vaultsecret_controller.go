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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	secretsbrokerv1alpha1 "github.com/gargath/secrets-broker/api/v1alpha1"
)

// VaultSecretReconciler reconciles a VaultSecret object
type VaultSecretReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=secretsbroker.phil.pub,resources=vaultsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secretsbroker.phil.pub,resources=vaultsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=secrets,verbs=get;list;watch;create;update;patch
func (r *VaultSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaultsecret", req.NamespacedName)

	log.Info(fmt.Sprintf("Reconciling: %+v", req))

	var vs secretsbrokerv1alpha1.VaultSecret
	if err := r.Get(ctx, req.NamespacedName, &vs); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Unable to fetch VaultSecret %s", req.NamespacedName))
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info(fmt.Sprintf("retrieved VaultSecret: %s", req.NamespacedName))
	log.Info(fmt.Sprintf("Status on this thing is: %+v", vs.Status))
	switch vs.Status.Phase {
	case "":
		myVs := vs.DeepCopy()
		myVs.Status = secretsbrokerv1alpha1.VaultSecretStatus{
			Phase: "Pending",
		}
		if err := r.Update(ctx, myVs); err != nil {
			log.Error(err, "unable to update VaultSecret status")
			return ctrl.Result{}, err
		}
	case "InSync":
		log.Info("VaultSecret is InSync, later on we'll check the source, for now nothing to do")
		return ctrl.Result{}, nil
	case "Pending":
		var s v1.Secret
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: req.Namespace,
			Name:      req.Name,
		}, &s)
		if err != nil {
			if apierrors.IsNotFound(err) {
				s = v1.Secret{}
				s.Namespace = req.Namespace
				s.Name = req.Name
				s.Data = make(map[string][]byte)
				s.Type = vs.Spec.Spec.Type
				log.Info(fmt.Sprintf("Now iterating over these fieldrefs: %+v", vs.Spec.Spec.FieldRefs))
				for field, source := range vs.Spec.Spec.FieldRefs {
					log.Info(fmt.Sprintf("Handling field %s, mapping to %s", field, source))
					s.Data[field] = []byte(fmt.Sprintf("handled-%s", source))
				}
				controllerutil.SetControllerReference(&vs, &s, r.Scheme)
				err := r.Client.Create(ctx, &s)
				if err != nil {
					log.Error(err, "failed to create secret for VaultSecret")
					return ctrl.Result{}, err
				}
				log.Info("Done handling the Secret, now updating the Phrase")
				myVs := vs.DeepCopy()
				myVs.Status.Phase = "InSync"
				apimeta.SetStatusCondition(&myVs.Status.Conditions, metav1.Condition{
					Type:               "StaleSecret",
					Status:             "False",
					Reason:             "SecretSynchronized",
					Message:            "The managed Secret is up to date with the Vault source",
					LastTransitionTime: metav1.NewTime(time.Now()),
				})
				err = r.Client.Update(ctx, myVs)
				if err != nil {
					log.Error(err, "failed up update phase")
				}
				r.Recorder.Event(myVs, v1.EventTypeNormal, "SecretCreated", "The managed Secret has been created")

				log.Info("Phase updated")
			} else {
				log.Error(err, "failed to get corresponding Secret")
			}
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Info(fmt.Sprintf("Found Secret: %+v", s))

	default:
		err := fmt.Errorf("unknown phase")
		log.Error(err, fmt.Sprintf("Phase %s on VaultSecret %s not handled by this controller", vs.Status.Phase, req.NamespacedName))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *VaultSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsbrokerv1alpha1.VaultSecret{}).
		Complete(r)
}
