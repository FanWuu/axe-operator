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

package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/presslabs/controller-util/pkg/meta"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	databasev1 "axe/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MysqlReconciler reconciles a Mysql object
type MysqlReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var FinalizerName = "axe-finalizer"

//+kubebuilder:rbac:groups=database.wufan,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database.wufan,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database.wufan,resources=mysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps;secrets;services;pods;pods/exec;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;create;patch
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mysql object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified bys
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *MysqlReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	ins := &databasev1.Mysql{}
	if err := r.Get(ctx, req.NamespacedName, ins); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add finalizer if is not added on the resource.
	if !meta.HasFinalizer(&ins.ObjectMeta, FinalizerName) {
		meta.AddFinalizer(&ins.ObjectMeta, FinalizerName)
		if err := r.Update(ctx, ins); err != nil {
			return ctrl.Result{}, err
		}
	}

	// delete cr relation resource 例如 StatefulSet、Headless Service 等
	if !ins.GetDeletionTimestamp().IsZero() {
		log.Log.Info("mysql cluster is deleting", "clusterspace", ins.Namespace, "clustername", ins.Name)
		if err := r.cleanupRelatedResources(ctx, ins); err != nil {
			return ctrl.Result{}, err
		}
		log.Log.Info("cleanup crd sucess ")
		return ctrl.Result{}, nil
	}

	// apply resources
	if _, err := ApplyResources(ctx, r.Client, ins); err != nil {
		log.Log.Error(err, "Apply Resources failed ")
		return ctrl.Result{}, err
	}

	// // create cluster
	if _, err := CreateCluster(ctx, r.Client, ins); err != nil {
		log.Log.Error(err, "create cluster failed ")
		return ctrl.Result{}, err
	}

	// update lables
	//fix the error : "the object has been modified; please apply your changes to the latest version and try again"
	statefulSet := &appsv1.StatefulSet{}
	r.Get(ctx, req.NamespacedName, statefulSet)
	if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas &&
		statefulSet.Status.CurrentRevision == statefulSet.Status.UpdateRevision &&
		statefulSet.ObjectMeta.Labels["clusterstatus"] == databasev1.MgrNOTinstalled {
		statefulSet.ObjectMeta.Labels["clusterstatus"] = databasev1.Mgrinstalled

		if err := r.Update(ctx, statefulSet); err != nil {
			log.Log.Error(err, "update statefulset lables failed")
			return ctrl.Result{}, err
		}
		time.Sleep(3 * time.Second)
		log.Log.Info("update statefulset lable MGR_INSTALLED")
	}

	// create router deployment
	if _, err := CreateRouter(ctx, r.Client, ins); err != nil {
		log.Log.Error(err, "create router failed ")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Mysql{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Complete(r)
}
