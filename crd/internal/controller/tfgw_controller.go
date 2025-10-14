/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	ingressv1 "github.com/codescalers/kubecloud/crd/api/v1"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// finalizer string set to delay the delation until external resource cleanup
const tfgwFinalizer = "finalizer.tfgw.ingress.grid.tf"

// TFGWReconciler reconciles a TFGW object
type TFGWReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ingress.grid.tf,resources=tfgws,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ingress.grid.tf,resources=tfgws/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ingress.grid.tf,resources=tfgws/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TFGW object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *TFGWReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var tfgw ingressv1.TFGW
	if err := r.Get(ctx, req.NamespacedName, &tfgw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// updating status trigger another reconcile, avoid by checking the status
	if meta.IsStatusConditionTrue(tfgw.Status.Conditions, ingressv1.ConditionTypeReady) {
		return ctrl.Result{}, nil
	}

	mne := os.Getenv("MNEMONIC")
	net := os.Getenv("NETWORK")
	if net == "" || mne == "" {
		klog.Warning("ThreeFold network or mnemonic not configured, skipping gateway deployment")

		meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
			Type:    ingressv1.ConditionTypeError,
			Status:  metav1.ConditionFalse,
			Reason:  "MissingConfiguration",
			Message: "ThreeFold network or mnemonic not configured",
		})

		tfgw.Status.Message = "Configuration missing: MNEMONIC or NETWORK not set"
		_ = r.Status().Update(ctx, &tfgw)

		return ctrl.Result{}, nil
	}

	token := os.Getenv("TOKEN")
	if token != "" {
		decryptedMne, err := decrypt(token, mne)
		if err != nil {
			klog.Warningf("Failed to decrypt mnemonic, using raw value: %v", err)

			meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
				Type:    ingressv1.ConditionTypeError,
				Status:  metav1.ConditionFalse,
				Reason:  "DecryptionFailed",
				Message: "Failed to decrypt mnemonic, using raw value",
			})

			tfgw.Status.Message = "Failed to decrypt mnemonic, using raw value"
			_ = r.Status().Update(ctx, &tfgw)

			return ctrl.Result{}, nil
		}

		mne = decryptedMne
	}

	sessionID, err := generateSessionId()
	if err != nil {
		klog.Warning("Failed to generate session id")

		meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
			Type:    ingressv1.ConditionTypeError,
			Status:  metav1.ConditionFalse,
			Reason:  "SessionGenerationFailed",
			Message: "Failed to generate session ID",
		})

		tfgw.Status.Message = "Failed to generate session ID"
		_ = r.Status().Update(ctx, &tfgw)

		return ctrl.Result{}, nil
	}

	pluginClient, err := deployer.NewTFPluginClient(
		mne,
		deployer.WithNetwork(net),
		deployer.WithDisableSentry(),
		deployer.WithSessionId(sessionID),
	)
	if err != nil {
		meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
			Type:    ingressv1.ConditionTypeError,
			Status:  metav1.ConditionFalse,
			Reason:  "ClientCreationFailed",
			Message: fmt.Sprintf("Failed to create TF plugin client: %v", err),
		})

		tfgw.Status.Message = "Failed to create TF plugin client"
		_ = r.Status().Update(ctx, &tfgw)

		return ctrl.Result{}, fmt.Errorf("failed to create TF plugin client: %w", err)
	}
	defer pluginClient.Close()

	// Handle deletion logic
	if !tfgw.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&tfgw, tfgwFinalizer) {
			log := logf.FromContext(ctx)
			log.Info("Cleaning up before deletion", "host", tfgw.Spec.Hostname)

			if err := pluginClient.CancelByProjectName(tfgw.Spec.Hostname); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to cancel gateway: %w", err)
			}

			controllerutil.RemoveFinalizer(&tfgw, tfgwFinalizer)
			if err := r.Update(ctx, &tfgw); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&tfgw, tfgwFinalizer) {
		controllerutil.AddFinalizer(&tfgw, tfgwFinalizer)
		if err := r.Update(ctx, &tfgw); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle creation logic
	res, err := deployGateway(pluginClient, GWRequest{
		Hostname: tfgw.Spec.Hostname,
		Backends: tfgw.Spec.Backends,
	})
	if err != nil {
		meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
			Type:    ingressv1.ConditionTypeError,
			Status:  metav1.ConditionFalse,
			Reason:  "DeploymentFailed",
			Message: fmt.Sprintf("Failed to deploy gateway: %v", err),
		})

		tfgw.Status.Message = "Failed to create DNS record"
		_ = r.Status().Update(ctx, &tfgw)

		return ctrl.Result{}, fmt.Errorf("failed to deploy gateway: %w", err)
	}

	tfgw.Status.FQDN = res.FQDN
	tfgw.Status.Message = "DNS record created"

	meta.SetStatusCondition(&tfgw.Status.Conditions, metav1.Condition{
		Type:    ingressv1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "Ready",
		Message: "Gateway is ready and accessible",
	})
	_ = r.Status().Update(ctx, &tfgw)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TFGWReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1.TFGW{}).
		Named("tfgw").
		Complete(r)
}
