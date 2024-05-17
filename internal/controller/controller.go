package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/presslabs/controller-util/pkg/meta"
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
	statefulSetName := fmt.Sprintf(ins.Name)
	statefulSet := syncer.MysqlStatefulset(ins)
	if err := r.Get(ctx, types.NamespacedName{Name: statefulSetName, Namespace: ins.Namespace}, statefulSet); err == nil {
		if err := r.Delete(ctx, statefulSet); err != nil {
			return fmt.Errorf("failed to delete StatefulSet %s: %w", statefulSetName, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get StatefulSet %s: %w", statefulSetName, err)
	}

	// cleanup deployment
	// deploymentname := fmt.Sprintf(ins.Name)
	// deployment := syncer.RouterDeployment(ins)
	// if err := r.Get(ctx, types.NamespacedName{Name: deploymentname, Namespace: ins.Namespace}, deployment); err == nil {
	// 	if err := r.Delete(ctx, deployment); err != nil {
	// 		return fmt.Errorf("failed to delete deployment %s: %w", deploymentname, err)
	// 	}
	// } else if !errors.IsNotFound(err) {
	// 	return fmt.Errorf("failed to get deployment %s: %w", deploymentname, err)
	// }

	// cleanup Headless Service
	svcName := fmt.Sprintf(ins.Name)
	svc := syncer.MysqlHeadlesSVC(ins)
	if err := r.Get(ctx, types.NamespacedName{Name: svcName, Namespace: ins.Namespace}, svc); err == nil {
		if err := r.Delete(ctx, svc); err != nil {
			return fmt.Errorf("failed to delete Service %s: %w", svcName, err)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service %s: %w", svcName, err)
	}

	// cleanup configmap
	configname := fmt.Sprintf("%s-%s", ins.Name, "mysql")
	configmap := syncer.MysqlConfigmap(ins)
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

	key := client.ObjectKeyFromObject(obj)

	// Check if the resource already exists
	err := c.Get(ctx, key, obj)
	switch {
	case err == nil:
		// Resource exists, update it
		return c.Update(ctx, obj)
	case apierrors.IsNotFound(err):
		// Resource doesn't exist, create it
		return c.Create(ctx, obj)
	default:
		return fmt.Errorf("failed to get existing resource: %w", err)
	}
}

func CreatCluster(ctx context.Context, r client.Client, ins *databasev1.Mysql) (ctrl.Result, error) {

	_ = log.FromContext(ctx)

	if err := CreateOrUpdate(ctx, r, syncer.MysqlHeadlesSVC(ins)); err != nil {
		return ctrl.Result{}, err
	}
	if err := CreateOrUpdate(ctx, r, syncer.MysqlConfigmap(ins)); err != nil {
		return ctrl.Result{}, err
	}
	if err := CreateOrUpdate(ctx, r, syncer.MysqlStatefulset(ins)); err != nil {
		return ctrl.Result{}, err
	}

	//create innodb cluster
	statefulSetName := fmt.Sprintf(ins.Name)
	statefulSet := syncer.MysqlStatefulset(ins)
	if err := r.Get(ctx, types.NamespacedName{Name: statefulSetName, Namespace: ins.Namespace}, statefulSet); err == nil {

		// 检查 StatefulSet 是否正常运行
		if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas && statefulSet.ObjectMeta.Labels["clusterstatus"] == "MGR_NOT_INSTALLED" {
			// dba.createcluster()
			log.Log.Info("StatefulSet is running and innodb cluster lables MGR_NOT_INSTALLED")
			if err := syncer.CreateOrUpdateMGR(ctx, ins); err == nil {
				log.Log.Info("CreateOrUpdateMGR  SUCCESS")

				//  if CreateOrUpdateMGR success r,update.statefulset lable clusterstatys
				statefulSet.ObjectMeta.Labels["clusterstatus"] = "MGR_INSTALLED"

				// 创建一个 JSON 补丁
				patchBytes, err := json.Marshal(statefulSet)
				if err != nil {
					return ctrl.Result{}, err
				}

				// 创建一个 RawPatch 实例
				rawPatch := client.RawPatch(types.MergePatchType, patchBytes)

				// 使用 RawPatch 更新 StatefulSet
				if err := r.Patch(ctx, statefulSet, rawPatch); err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, nil

			} else {
				log.Log.Error(err, "CreateOrUpdateMGR  FAILED")

				statefulSet.ObjectMeta.Labels["clusterstatus"] = "MGR_INSTALLED_FAILED"

				// 创建一个 JSON 补丁
				patchBytes, err := json.Marshal(statefulSet)
				if err != nil {
					return ctrl.Result{}, err
				}

				// 创建一个 RawPatch 实例
				rawPatch := client.RawPatch(types.MergePatchType, patchBytes)

				// 使用 RawPatch 更新 StatefulSet
				if err := r.Patch(ctx, statefulSet, rawPatch); err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, nil

			}

		} else if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas && statefulSet.ObjectMeta.Labels["clusterstatus"] == "MGR_INSTALLED" {
			// innodb cluster has already installed
			log.Log.Info("StatefulSet is running and innodb cluster lables MGR_INSTALLED")
			return ctrl.Result{}, nil

		} else {
			log.Log.Error(err, "StatefulSet is not running normally ", statefulSet)
			return ctrl.Result{}, err

		}

	} else if !errors.IsNotFound(err) {
		log.Log.Error(err, "statefulset: ", statefulSetName, "not fondun")
	}
	return ctrl.Result{}, nil
}
