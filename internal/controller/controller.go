package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/presslabs/controller-util/pkg/meta"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "axe/api/v1"
	syncer "axe/syncer/mysql"
)

func (r *MysqlReconciler) cleanupRelatedResources(ctx context.Context, ins *databasev1.Mysql) error {

	if ins == nil {
		return fmt.Errorf("object to create or update must not be nil")
	}

	defer func() {
		// 如果清理成功，删除 Finalizer（如果有）
		meta.RemoveFinalizer(&ins.ObjectMeta, FinalizerName)
		if err := r.Update(ctx, ins); err != nil {
			log.Log.Error(err, "failed to update cluster")
		}
	}()
	// cleanup StatefulSet
	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(ins), statefulSet); err == nil {
		if err := r.Delete(ctx, statefulSet); err != nil {
			return fmt.Errorf("failed to delete StatefulSet %s: %w", ins.Name, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get StatefulSet %s: %w", ins.Name, err)
	}

	// cleanup Headless Service
	svc := &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(ins), svc); err == nil {
		if err := r.Delete(ctx, svc); err != nil {
			return fmt.Errorf("failed to delete Service %s: %w", ins.Name, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service %s: %w", ins.Name, err)
	}

	// cleanup router Service
	svc = &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: ins.Name + "-router", Namespace: ins.Namespace}, svc); err == nil {
		if err := r.Delete(ctx, svc); err != nil {
			return fmt.Errorf("failed to delete Service %s: %w", ins.Name, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service %s: %w", ins.Name, err)
	}

	// cleanup router Service
	svc = &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: ins.Name + "-router-node", Namespace: ins.Namespace}, svc); err == nil {
		if err := r.Delete(ctx, svc); err != nil {
			return fmt.Errorf("failed to delete Service %s: %w", ins.Name, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service %s: %w", ins.Name, err)
	}

	// cleanup deployment
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: ins.Name + "-router", Namespace: ins.Namespace}, deploy); err == nil {
		if err := r.Delete(ctx, deploy); err != nil {
			return fmt.Errorf("failed to delete Service %s: %w", ins.Name, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service %s: %w", ins.Name, err)
	}

	// cleanup configmap
	configname := fmt.Sprintf("%s-%s", ins.Name, "mysql")
	configmap := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: configname, Namespace: ins.Namespace}, configmap); err == nil {
		if err := r.Delete(ctx, configmap); err != nil {
			return fmt.Errorf("failed to delete configmap %s: %w", configmap, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get configmap %s: %w", configmap, err)
	}

	return nil
}

// CreateOrUpdate performs a create-or-update operation on the given object.
// If the object does not exist, it is created. If it already exists, it is updated.
func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object) error {
	if obj == nil {
		return fmt.Errorf("object to create or update must not be nil")
	}
	// Check if the resource already exists
	existingObj := obj.DeepCopyObject().(client.Object)
	err := c.Get(ctx, client.ObjectKeyFromObject(obj), existingObj)

	switch {
	case err == nil:
		// Resource exists, check for updates before updating
		if equality.Semantic.DeepEqual(existingObj, obj) {
			log.Log.Info("No changes detected, skipping update", "namespace", obj.GetNamespace(), "kind", obj.GetObjectKind(), "name", obj.GetName())
			return nil
		}
		log.Log.Info("update resource due to changes", "objspeace", obj.GetNamespace(), "objtype", obj.GetObjectKind(), "objname", obj.GetName())
		return c.Update(ctx, obj)
	case apierrors.IsNotFound(err):
		// Resource doesn't exist, create it
		log.Log.Info("create resource", "objspeace", obj.GetNamespace(), "objtype", obj.GetObjectKind(), "objname", obj.GetName())
		return c.Create(ctx, obj)
	default:
		return fmt.Errorf("failed to get existing resource: %w", err)
	}
}

func ApplyResources(ctx context.Context, r client.Client, ins *databasev1.Mysql) (ctrl.Result, error) {
	log.Log.Info("create or update resource", "clusterspace", ins.Namespace, "clustername", ins.Name)

	if err := CreateOrUpdate(ctx, r, syncer.MysqlHeadlesSVC(ins)); err != nil {
		return ctrl.Result{}, err
	}

	if err := CreateOrUpdate(ctx, r, syncer.MysqlConfigmap(ins)); err != nil {
		return ctrl.Result{}, err
	}

	if err := CreateOrUpdate(ctx, r, syncer.RouterConfigmap(ins)); err != nil {
		return ctrl.Result{}, err
	}

	if err := CreateOrUpdate(ctx, r, syncer.MysqlStatefulset(ins)); err != nil {
		return ctrl.Result{}, err
	}

	log.Log.Info("Apply Resources sucess ")
	return ctrl.Result{}, nil
}

func CreateRouter(ctx context.Context, r client.Client, ins *databasev1.Mysql) (ctrl.Result, error) {
	log.Log.Info("create  router resource", "clusterspace", ins.Namespace, "clustername", ins.Name)

	if err := CreateOrUpdate(ctx, r, syncer.RouterDeployment(ins)); err != nil {
		return ctrl.Result{}, err
	}
	if err := CreateOrUpdate(ctx, r, syncer.RouterClusterSVC(ins)); err != nil {
		return ctrl.Result{}, err
	}
	if err := CreateOrUpdate(ctx, r, syncer.RouterNodeSVC(ins)); err != nil {
		return ctrl.Result{}, err
	}
	log.Log.Info("Create Routers sucess ")
	return ctrl.Result{}, nil
}
func CreateCluster(ctx context.Context, r client.Client, ins *databasev1.Mysql) (ctrl.Result, error) {
	//create innodb cluster
	statefulSet := &appsv1.StatefulSet{}
	time.Sleep(10 * time.Second)

	if err := r.Get(ctx, client.ObjectKeyFromObject(ins), statefulSet); err == nil {

		// 检查 StatefulSet 是否正常运行
		if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas &&
			statefulSet.Status.CurrentRevision == statefulSet.Status.UpdateRevision &&
			statefulSet.ObjectMeta.Labels["clusterstatus"] == "MGR_NOT_INSTALLED" {
			// dba.createcluster()
			log.Log.Info("StatefulSet is running and innodb cluster lables MGR_NOT_INSTALLED")
			if err := syncer.CreateMGR(ctx, ins); err == nil {
				log.Log.Info("Create innodb cluster SUCCESS")

				return ctrl.Result{}, nil
			} else {
				log.Log.Error(err, "Create innodb cluster FAILED")
				return ctrl.Result{}, err
			}

		} else if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas &&
			statefulSet.Status.CurrentRevision == statefulSet.Status.UpdateRevision &&
			statefulSet.ObjectMeta.Labels["clusterstatus"] == "MGR_INSTALLED" {
			// innodb cluster has already installed
			log.Log.Info("StatefulSet is running and innodb cluster lables MGR_INSTALLED")
			return ctrl.Result{}, nil

		} else {
			log.Log.Error(err, "StatefulSet is not running normally ")
			log.Log.Info("StatefulSet is not running normally", "ReadyReplicas:", statefulSet.Status.ReadyReplicas, "Replicas", statefulSet.Status.Replicas)
			log.Log.Info("StatefulSet is not running normally", "CurrentRevision:", statefulSet.Status.CurrentRevision, "UpdateRevision", statefulSet.Status.UpdateRevision)
			log.Log.Info("StatefulSet is not running normally", "clusterstatus:", statefulSet.ObjectMeta.Labels["clusterstatus"])

			return ctrl.Result{}, err
		}

	} else if !errors.IsNotFound(err) {
		log.Log.Error(err, "statefulset: ", ins.Name, "not fondun")
	}
	return ctrl.Result{}, nil
}
