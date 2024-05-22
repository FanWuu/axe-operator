package syncer

import (
	databasev1 "axe/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//  docker run \
//   -e MYSQL_HOST=localhost \
//   -e MYSQL_PORT=3306 \
//   -e MYSQL_USER=mysql \
//   -e MYSQL_PASSWORD=mysql \
//   -e MYSQL_INNODB_CLUSTER_MEMBERS=3 \
//   -e MYSQL_ROUTER_BOOTSTRAP_EXTRA_OPTIONS="--conf-use-socket --conf-use-gr-notification"
//   -ti container-registry.oracle.com/mysql/community-router

//	func clusterHost(ins *databasev1.Mysql) string {
//		clusterHost := ""
//		// for i := 0; i < int(ins.Spec.Replica); i++ {
//		// 	clusterHost = clusterHost + ins.Name + "-" + strconv.Itoa(i) + "." + ins.Name + "." + ins.Namespace + ".svc.cluster.local:3306:"
//		// }
//		clusterHost = clusterHost + ins.Name + "-" + strconv.Itoa(0) + "." + ins.Name + "." + ins.Namespace + ".svc.cluster.local"
//		return clusterHost
//	}
func Routercontainer(ins *databasev1.Mysql) []corev1.Container {
	return []corev1.Container{
		{
			Name:            ins.Name + "-router",
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
					Value: ins.Name + "." + ins.Namespace + ".svc.cluster.local",
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
					Name:  "MYSQL_CREATE_ROUTER_USER",
					Value: "0",
				},
				{
					Name:  "MYSQL_PASSWORD",
					Value: ins.Spec.Mysql.RootPassword,
				},
				{
					Name:  "MYSQL_INNODB_CLUSTER_MEMBERS",
					Value: "3",
				},
				{
					//https://dev.mysql.com/doc/mysql-router/8.3/en/mysql-router-installation-docker.html
					//https://github.com/mysql/mysql-operator/blob/trunk/mysqloperator/controller/innodbcluster/router_objects.py
					Name:  "MYSQL_ROUTER_BOOTSTRAP_EXTRA_OPTIONS",
					Value: "--conf-set-option=DEFAULT.unknown_config_option=warning --conf-set-option=DEFAULT.max_total_connections=10240 ",
				},
				// 添加其他必要的环境变量
			},
		},
	}
}

// 也可以其多个服务，独立提供访问
func RouterDeployment(ins *databasev1.Mysql) *appsv1.Deployment {
	if ins == nil || ins.Spec.Replica < 0 {
		// 在实际场景中，应该处理这个错误，比如返回一个错误或记录日志
		return nil
	}
	replicas := ins.Spec.Router.Replica

	RouterDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-router",
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
