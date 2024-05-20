package syncer

import (
	databasev1 "axe/api/v1"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func env(ins *databasev1.Mysql) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "MYSQL_ROOT_PASSWORD",
			Value: ins.Spec.Mysql.RootPassword,
		},
		{
			Name:  "NAMESPACE",
			Value: ins.Namespace,
		},
		{
			Name:  "REPLICA",
			Value: strconv.Itoa(int(ins.Spec.Replica)),
		},
		{
			Name:  "SERVICE_NAME",
			Value: ins.Name,
		},
	}
}

func InitContainers(ins *databasev1.Mysql) []corev1.Container {
	return []corev1.Container{
		{
			Name:  "init-mysql",
			Image: ins.Spec.Mysql.MysqlImage,
			Command: []string{
				"sh",
				"-c",
				`
				# 解析 HOSTNAME 获取 Pod 索引
				POD_INDEX=$(echo $HOSTNAME | awk -F'-' '{print $NF}')
				TIMEUNIX=$(date +%s | awk '{print substr($0,length()-4)}')
				# 将索引写入到 MySQL 配置文件（假设为 /etc/mysql/conf.d/server-id.cnf）
				echo "[mysqld]" > /etc/mysql/conf.d/server-id.cnf
				echo "server-id=$TIMEUNIX$POD_INDEX" >> /etc/mysql/conf.d/server-id.cnf
				#mysql-axe-2.mysql-axe.default.svc.cluster.local mysql-axe-2
				echo "report_host=$HOSTNAME.$SERVICE_NAME.$NAMESPACE.svc.cluster.local" >> /etc/mysql/conf.d/server-id.cnf

				ln -sf /mnt/config/* /etc/mysql/conf.d/
				`,
			},
			Env: env(ins),
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "server-id",
					MountPath: "/etc/mysql/conf.d/",
				},
				{
					Name:      ins.Name + "-mysql",
					MountPath: "/mnt/config/",
				},
			},
		},
	}
}

func mysqlContainers(ins *databasev1.Mysql) []corev1.Container {
	return []corev1.Container{
		{
			Name:            "mysql",
			Image:           ins.Spec.Mysql.MysqlImage,
			ImagePullPolicy: ins.Spec.PodPolicy.ImagePullPolicy,

			Ports: []corev1.ContainerPort{
				{
					Name:          "mysql",
					ContainerPort: 3306,
				},
				{
					Name:          "mysqlx",
					ContainerPort: 33060,
				},
				{
					Name:          "gr-xcom",
					ContainerPort: 33061,
				},
			},
			Env: env(ins),
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "server-id",
					MountPath: "/etc/mysql/conf.d/",
				},
				{
					Name:      ins.Name + "-mysql",
					MountPath: "/mnt/config/",
				},
				{
					Name:      "mysql-data",
					MountPath: "/var/lib/mysql",
				},
			},
			Resources: ins.Spec.PodPolicy.ExtraResources,
		},
	}
}

func MysqlStatefulset(ins *databasev1.Mysql) *appsv1.StatefulSet {
	if ins == nil || ins.Spec.Replica < 0 {
		// 在实际场景中，应该处理这个错误，比如返回一个错误或记录日志
		return nil
	}

	lables := map[string]string{
		"clustername": ins.Name,
		"app":         MYSQLAPP,
	}

	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name,
			Namespace: ins.Namespace,
			Labels: map[string]string{
				"clustername":   ins.Name,
				"app":           MYSQLAPP,
				"clusterstatus": MgrNOTinstalled,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &ins.Spec.Replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: lables,
			},
			ServiceName: ins.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lables,
				},

				Spec: corev1.PodSpec{
					InitContainers: InitContainers(ins),
					Containers:     mysqlContainers(ins),
					// 添加其他Pod配置，如NodeSelector、Affinity、Tolerations等
					Volumes: []corev1.Volume{

						{
							Name: ins.Name + "-mysql",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ins.Name + "-mysql",
									},
								},
							},
						},
						{
							Name: "server-id",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{

							Name: "mysql-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/data/mysql/" + ins.Namespace + "/" + ins.Name,
									// Type: &corev1.HostPathDirectoryOrCreate,
								},
							},
						},
					},
				},
			},
			// 添加 VolumeClaimTemplates 如果需要
		},
	}

	return statefulSet
}
