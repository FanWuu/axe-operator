package innodbcluster

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

func ExeCmd(cmd string) (string, error) {
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
		return stderr.String(), err
	}
	log.Log.Info("Result: " + out.String())
	return out.String(), nil
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
			log.Log.Info("Ping attempt success ", "host:", host)
			return true
		}
		// 如果发生错误，打印错误信息并等待一段时间重试
		log.Log.Info("Ping attempt failed try latter ", "host:", host)
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
			log.Log.Info("mysql is not ready", "host", host)
			return fmt.Errorf("mysql is not ready")
		}

		if i == 0 {
			// if cluster status is ok return nil
			cmd := `/usr/bin/mysqlsh -uroot -p` + passwd + ` -h` + host + ` --cluster  -e "print(cluster.status())"`
			if _, err := ExeCmd(cmd); err == nil {
				log.Log.Info("cluster is ready")
				return nil
			}
			// else create cluster
			log.Log.Info("create innodb cluster", "host", host)

			cmd = `echo Y |/usr/bin/mysqlsh -uroot -p` + passwd + ` -h` + host + ` -e "dba.createCluster('mgr')"`
			ExeCmd(cmd)
		} else {
			// 添加节点

			log.Log.Info("add instance to cluster", "host", host)
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
