package syncer

import (
	databasev1 "axe/api/v1"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// TODO
// 方案一  kubectl exec -it mysql-0 -n mysql -- mysql -uroot -p  方式初始化集群
// 方案二  使用initcontainer方式初始化集群 需要自定义一个container ，在mysql 镜像的基础上打包sider代码
// 代码逻辑 1. 判断pod index 如果自己是最后一个 则去前面的节点创建集群，完成后退出
// 问题的关键点  最后一个节点如何在initcontainer 退出前加入集群

// 方案三  operator manager container 中集成 mysql shell 直接本地调用mysqlsh createcluster , addinstance

// 1. 检查statefulset.status == running && lables.clusterstatus== uninit pod数量大于等于3 则执行创建集群操作
// 2. 创建集群操作
// 3. 添加实例操作
// 4. 删除实例操作
// 5. 扩容操作
// 6. 缩容操作
// 7. 升级操作
// 完成以上操作后修改标签状态
func ExeCmd(cmd string) (error, string) {
	//TODO CHECK ERROR RESULT
	c := exec.Command("sh", "-c", cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()
	log.Log.Info("exec", " cmd :", cmd)
	if err != nil {
		log.Log.Info(fmt.Sprint(err) + ": " + stderr.String())
		return err, stderr.String()
	}
	log.Log.Info("Result: " + out.String())
	return nil, out.String()
}

func pingMySQ(host string, passwd string) bool {
	db, err := sql.Open("mysql", `root:`+passwd+`@tcp(`+host+`:3306)/mysql?charset=utf8mb4`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	attempt := 0
	for attempt < 50 {
		err := db.PingContext(context.Background())
		if err == nil {
			log.Log.Info("Ping attempt success ")
			return true
		}
		// 如果发生错误，打印错误信息并等待一段时间重试
		log.Log.Info("Ping attempt failed try latter ")
		time.Sleep(time.Second * 2)
		attempt++
	}
	// 达到最大尝试次数，返回 false
	log.Log.Info("Ping attempt failed ")

	return false
}
func CreateMGR(ctx context.Context, ins *databasev1.Mysql) error {

	//mysql-axe-2.mysql-axe.default.svc.cluster.local
	host0 := ins.Name + "-" + strconv.Itoa(0) + "." + ins.Name + "." + ins.Namespace + ".svc.cluster.local"
	passwd := ins.Spec.Mysql.RootPassword

	for i := 0; i < int(ins.Spec.Replica); i++ {
		time.Sleep(time.Second * 3)
		host := ins.Name + "-" + strconv.Itoa(i) + "." + ins.Name + "." + ins.Namespace + ".svc.cluster.local"

		if !pingMySQ(host, passwd) {
			// TODO
			log.Log.Info("mysql is not ready", "host:", host)
			return fmt.Errorf("mysql is not ready")
		}

		if i == 0 {
			// TODO  if cluster status is ok return nil
			cmd := `/usr/bin/mysqlsh -uroot -p` + passwd + ` -h` + host + ` --cluster  -e "print(cluster.status())"`
			if err, _ := ExeCmd(cmd); err == nil {
				log.Log.Info("cluster is ready")
				return nil
			}
			// else create cluster
			log.Log.Info("create innodb cluster", "host:", host)

			cmd = `echo Y |/usr/bin/mysqlsh -uroot -p` + passwd + ` -h` + host + ` -e "dba.createCluster('mgr')"`
			ExeCmd(cmd)
		} else {
			// 添加节点

			log.Log.Info("add instance to cluster", "host:", host)
			cmd := `echo C |/usr/bin/mysqlsh -uroot -p` + passwd + ` -h` + host0 + ` --cluster -e "cluster.addInstance('root@` + host + `:3306')"`
			ExeCmd(cmd)

		}
	}

	// cluster.rescan()
	cmd := `echo y |/usr/bin/mysqlsh -uroot -p` + ins.Spec.Mysql.RootPassword + ` -h` + host0 + ` --cluster  -e "cluster.rescan()"`
	ExeCmd(cmd)

	// print cluster.status()
	cmds := `/usr/bin/mysqlsh  -uroot -p` + ins.Spec.Mysql.RootPassword + ` -h` + host0 + `  --cluster  -e "print(cluster.status())"`
	ExeCmd(cmds)
	return nil
}
