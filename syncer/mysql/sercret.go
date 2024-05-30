package syncer

import (
	databasev1 "axe/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Mysqlsecret(ins *databasev1.Mysql) []corev1.Secret {
	return []corev1.Secret{
		{TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
			ObjectMeta: metav1.ObjectMeta{
				Name:      ins.Name,
				Namespace: ins.Namespace,
			},
			Data: map[string][]byte{
				"rootpasswd": []byte(ins.Spec.Mysql.RootPassword),
			},
		},
	}
}
