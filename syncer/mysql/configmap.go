package syncer

import (
	databasev1 "axe/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MysqlConfigmap(ins *databasev1.Mysql) *corev1.ConfigMap {

	conf := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-mysql",
			Namespace: ins.Namespace,
		},
		Data: map[string]string{
			"mysql.cnf":  mysqlConfigData,
			"plugin.cnf": PluginConfdata,
		},
	}
	return conf
}

func RouterConfigmap(ins *databasev1.Mysql) *corev1.ConfigMap {

	conf := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-router",
			Namespace: ins.Namespace,
		},
		Data: map[string]string{
			"router.cnf": RouterConf,
		},
	}
	return conf
}
