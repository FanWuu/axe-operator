package innodbcluster

import (
	databasev1 "axe/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MysqlHeadlesSVC(ins *databasev1.Mysql) *corev1.Service {
	labels := map[string]string{
		"clustername": ins.Name,
		"app":         databasev1.MYSQLAPP,
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
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
				{
					Name:     "mysqlx",
					Port:     33060,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "gr-xcom",
					Port:     33062,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func RouterClusterSVC(ins *databasev1.Mysql) *corev1.Service {
	labels := map[string]string{
		"clustername": ins.Name,
		"app":         databasev1.MYSQLROUTERAPP,
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-router",
			Namespace: ins.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "mysql-router-rw",
					Port:       6446,
					TargetPort: intstr.FromInt(6446),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "mysql-router-ro",
					Port:       6447,
					TargetPort: intstr.FromInt(6447),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func RouterNodeSVC(ins *databasev1.Mysql) *corev1.Service {
	labels := map[string]string{
		"clustername": ins.Name,
		"app":         databasev1.MYSQLROUTERAPP,
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-router-node",
			Namespace: ins.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "mysql-router-rw",
					Port:       31001,
					TargetPort: intstr.FromInt(6446),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "mysql-router-ro",
					Port:       31002,
					TargetPort: intstr.FromInt(6447),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}
