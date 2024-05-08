package syncer

// TODO
// 方案一  kubectl exec -it mysql-0 -n mysql -- mysql -uroot -p  方式初始化集群
// 方案二  使用initcontainer方式初始化集群 需要自定义一个container ，在mysql 镜像的基础上打包sider代码
//  代码逻辑 1. 判断pod index 如果自己是最后一个 则去前面的节点创建集群，完成后退出
// 问题的关键点  最后一个节点如何在initcontainer 退出前加入集群
