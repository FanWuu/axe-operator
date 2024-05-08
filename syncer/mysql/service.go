package syncer

import (
	databasev1 "axe/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MysqlHeadlesSVC(ins *databasev1.Mysql) *corev1.Service {
	labels := map[string]string{
		"clustername": ins.Name,
		"app":         "mysql",
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name,
			Namespace: ins.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "mysql",
					Port:     3306,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func MysqlClusterSVC(ins *databasev1.Mysql) *corev1.Service {
	labels := map[string]string{
		"clustername": ins.Name,
		"app":         "mysql",
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name,
			Namespace: ins.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "mysql",
					Port:     3306,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}
