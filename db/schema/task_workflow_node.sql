CREATE TABLE IF NOT EXISTS task_workflow_node (
  id BIGINT AUTO_INCREMENT NOT NULL,
  gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  workflow_id BIGINT NOT NULL COMMENT '关联工作流定义ID',
  node_id VARCHAR(64) NOT NULL COMMENT '节点唯一标识',
  node_type ENUM('ATOMIC', 'SUB_WORKFLOW', 'CONDITION', 'PARALLEL_SPLIT', 'PARALLEL_JOIN', 'DELAY', 'MANUAL_APPROVAL') NOT NULL COMMENT '节点类型',
  display_name VARCHAR(128) NOT NULL COMMENT '节点显示名称',
  task_atomic_def_id BIGINT DEFAULT NULL COMMENT '引用的原子任务ID',
  sub_workflow_id BIGINT DEFAULT NULL COMMENT '引用的子工作流ID',
  condition_expr TEXT DEFAULT NULL COMMENT '条件表达式',
  node_config JSON DEFAULT NULL COMMENT '节点额外配置',
  retry_policy JSON DEFAULT NULL COMMENT '节点级重试策略',
  timeout INT DEFAULT NULL COMMENT '节点超时时间',
  ord INT DEFAULT 0 COMMENT '执行顺序',
  PRIMARY KEY (id),
  UNIQUE KEY uk_wf_nid (workflow_id, node_id),
  KEY idx_wf_ord (workflow_id, ord)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作流节点定义表';
