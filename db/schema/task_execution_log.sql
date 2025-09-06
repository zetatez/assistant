CREATE TABLE IF NOT EXISTS task_execution_log (
  id BIGINT AUTO_INCREMENT NOT NULL,
  gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  workflow_instance_id BIGINT NOT NULL COMMENT '关联工作流实例ID',
  node_instance_id BIGINT DEFAULT NULL COMMENT '关联节点实例ID',
  task_atomic_instance_id BIGINT DEFAULT NULL COMMENT '关联原子任务实例ID',
  log_level ENUM('DEBUG', 'INFO', 'WARN', 'ERROR') NOT NULL DEFAULT 'INFO' COMMENT '日志级别',
  log_type ENUM('EXECUTION', 'ROLLBACK', 'RETRY', 'SYSTEM') NOT NULL DEFAULT 'EXECUTION' COMMENT '日志类型',
  message TEXT NOT NULL COMMENT '日志内容',
  context JSON DEFAULT NULL COMMENT '上下文信息',
  PRIMARY KEY (id),
  KEY idx_wi (workflow_instance_id),
  KEY idx_wi_ni (workflow_instance_id, node_instance_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表';
