package innodbcluster

var mysqlConfigData = `
[mysqld]
user    = mysql
port    = 3306
character-set-server = UTF8MB4
skip_name_resolve = 1
default_time_zone = "+8:00"

lock_wait_timeout = 3600
open_files_limit    = 65535
back_log = 1024
max_connections = 1024
max_connect_errors = 1000000
table_open_cache = 2048
table_definition_cache = 2048
thread_stack = 512K
sort_buffer_size = 4M
join_buffer_size = 4M
read_buffer_size = 8M
read_rnd_buffer_size = 4M
bulk_insert_buffer_size = 64M
thread_cache_size = 768
interactive_timeout = 600
wait_timeout = 600
tmp_table_size = 96M
max_heap_table_size = 96M

#log settings
log_timestamps = SYSTEM
log_error_verbosity = 3
slow_query_log = 1
log_slow_extra = 1
long_query_time = 0.01
log_queries_not_using_indexes = 1
log_throttle_queries_not_using_indexes = 60
min_examined_row_limit = 100
log_slow_admin_statements = 1
log_slow_slave_statements = 1
binlog_format = ROW
sync_binlog = 1
binlog_cache_size = 4M
max_binlog_cache_size = 6G
max_binlog_size = 1G
binlog_rows_query_log_events = 1
binlog_expire_logs_seconds = 604800
binlog_checksum = CRC32
gtid_mode = ON
enforce_gtid_consistency = TRUE

#replication settings
relay_log_recovery = 1
slave_parallel_type = LOGICAL_CLOCK
#并行复制线程数可以设置为逻辑CPU数量的2倍
slave_parallel_workers = 8
binlog_transaction_dependency_tracking = WRITESET
slave_preserve_commit_order = 1
slave_checkpoint_period = 2

#启用InnoDB并行查询优化功能
loose-force_parallel_execute = OFF
#设置每个SQL语句的并行查询最大并发度
loose-parallel_default_dop = 8
#设置系统中总的并行查询线程数，可以和最大逻辑CPU数量一样
loose-parallel_max_threads = 64
#并行执行时leader线程和worker线程使用的总内存大小上限，可以设置物理内存的5-10%左右
loose-parallel_memory_limit = 12G

#innodb settings
innodb_buffer_pool_size = 16G
innodb_buffer_pool_instances = 8
innodb_data_file_path = ibdata1:12M:autoextend
innodb_flush_log_at_trx_commit = 1
innodb_log_buffer_size = 32M
innodb_log_file_size = 2G
innodb_log_files_in_group = 3
innodb_doublewrite_files = 2
innodb_max_undo_log_size = 4G
# 根据您的服务器IOPS能力适当调整
# 一般配普通SSD盘的话，可以调整到 10000 - 20000
# 配置高端PCIe SSD卡的话，则可以调整的更高，比如 50000 - 80000
innodb_io_capacity = 4000
innodb_io_capacity_max = 8000
innodb_open_files = 65534
innodb_flush_method = O_DIRECT
innodb_lru_scan_depth = 4000
innodb_lock_wait_timeout = 10
innodb_rollback_on_timeout = 1
innodb_print_all_deadlocks = 1
innodb_online_alter_log_max_size = 4G
innodb_print_ddl_logs = 1
innodb_status_file = 1
innodb_status_output = 0
innodb_status_output_locks = 1
innodb_sort_buffer_size = 64M

`

var PluginConfdata = `
[mysqld]
loose-plugin_load_add = 'mysql_clone.so'
loose-plugin_load_add = 'group_replication.so'
loose-group_replication_start_on_boot = ON
loose-group_replication_bootstrap_group = OFF
loose-group_replication_exit_state_action = READ_ONLY
loose-group_replication_flow_control_mode = "DISABLED"
`

// mysqlrouter.conf
var RouterConf = `
[DEFAULT]
max_total_connections = 10240
logging_folder=
runtime_folder=/tmp/mysqlrouter/run
data_folder=/tmp/mysqlrouter/data
keyring_path=/tmp/mysqlrouter/data/keyring
master_key_path=/tmp/mysqlrouter/mysqlrouter.key
connect_timeout=5
read_timeout=30
dynamic_state=/tmp/mysqlrouter/data/state.json
client_ssl_cert=/tmp/mysqlrouter/data/router-cert.pem
client_ssl_key=/tmp/mysqlrouter/data/router-key.pem
client_ssl_mode=PREFERRED
server_ssl_mode=AS_CLIENT
server_ssl_verify=DISABLED
unknown_config_option=warning

[logger]
level=INFO

[metadata_cache:bootstrap]
cluster_type=gr
router_id=2
user=root
metadata_cluster=mgr
ttl=0.5
auth_cache_ttl=-1
auth_cache_refresh_interval=2
use_gr_notifications=0

[routing:bootstrap_rw]
bind_address=0.0.0.0
bind_port=6446
destinations=metadata-cache://mgr/?role=PRIMARY
routing_strategy=first-available
protocol=classic

[routing:bootstrap_ro]
bind_address=0.0.0.0
bind_port=6447
destinations=metadata-cache://mgr/?role=SECONDARY
routing_strategy=round-robin-with-fallback
protocol=classic

[routing:bootstrap_x_rw]
bind_address=0.0.0.0
bind_port=6448
destinations=metadata-cache://mgr/?role=PRIMARY
routing_strategy=first-available
protocol=x

[routing:bootstrap_x_ro]
bind_address=0.0.0.0
bind_port=6449
destinations=metadata-cache://mgr/?role=SECONDARY
routing_strategy=round-robin-with-fallback
protocol=x

[http_server]
port=8443
ssl=1
ssl_cert=/tmp/mysqlrouter/data/router-cert.pem
ssl_key=/tmp/mysqlrouter/data/router-key.pem

[http_auth_realm:default_auth_realm]
backend=default_auth_backend
method=basic
name=default_realm

[rest_router]
require_realm=default_auth_realm

[rest_api]

[http_auth_backend:default_auth_backend]
backend=metadata_cache

[rest_routing]
require_realm=default_auth_realm

[rest_metadata_cache]
require_realm=default_auth_realm

`
