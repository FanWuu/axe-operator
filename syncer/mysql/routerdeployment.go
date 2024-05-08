package syncer

import (
	databasev1 "axe/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Routercontainer(ins *databasev1.Mysql) []corev1.Container {
	return []corev1.Container{
		{
			Name:            "mysql-router",
			Image:           ins.Spec.Router.RouterImage,
			ImagePullPolicy: ins.Spec.PodPolicy.ImagePullPolicy,
			Ports: []corev1.ContainerPort{
				{
					Name:          "mysql-router",
					ContainerPort: 6446,
				},
			},
			// 设置必要的环境变量
			Env: []corev1.EnvVar{
				{
					Name:  "MYSQL_HOST",
					Value: "mysql-service", // 假设MySQL服务名为mysql-service
				},
				{
					Name:  "MYSQL_PORT",
					Value: "3306",
				},
				{
					Name:  "MYSQL_USER",
					Value: "root",
				},
				{
					Name:  "MYSQL_PASSWORD",
					Value: ins.Spec.Mysql.RootPassword, // 设置一个强密码，实际使用时应通过安全的方式传递
				},
				// 添加其他必要的环境变量
			},
		},
	}
}

func RouterDeployment(ins *databasev1.Mysql) *appsv1.Deployment {
	if ins == nil || ins.Spec.Replica < 0 {
		// 在实际场景中，应该处理这个错误，比如返回一个错误或记录日志
		return nil
	}
	replicas := ins.Spec.Router.Replica

	RouterDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-router",
			Namespace: ins.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas, // Use the replicas specified in the MysqlRouter CR
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"clustername": ins.Name,
					"app":         "mysql-router",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clustername": ins.Name,
						"app":         "mysql-router",
					},
				},
				Spec: corev1.PodSpec{
					Containers: Routercontainer(ins),
				},
			},
		},
	}
	return RouterDeployment
}
