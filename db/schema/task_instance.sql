
CREATE TABLE IF NOT EXISTS task_instance (
  id BIGINT AUTO_INCREMENT NOT NULL,
  gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  task_def_id BIGINT NOT NULL,
  parent_instance_id BIGINT DEFAULT NULL,
  ord INT DEFAULT 0,
  status ENUM('PENDING','RUNNING','SUCCESS','FAILED','CANCELLED') DEFAULT 'PENDING',
  result JSON DEFAULT NULL,
  err_msg TEXT DEFAULT NULL,
  FOREIGN KEY (task_def_id) REFERENCES task_def(id),
  FOREIGN KEY (parent_instance_id) REFERENCES task_instance(id) ON DELETE CASCADE,
  PRIMARY KEY (id),
  INDEX idx_pii_s (parent_instance_id, seq)
) COMMENT='任务实例表';

-- CREATE TABLE IF NOT EXISTS task_instance (
--   id BIGINT AUTO_INCREMENT NOT NULL,
--   gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--   gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--   task_def_id BIGINT NOT NULL COMMENT '关联的任务定义ID',
--   parent_instance_id BIGINT DEFAULT NULL COMMENT '父实例ID',
--   ord INT DEFAULT 0 COMMENT '同级实例顺序',
--   status ENUM('PENDING','RUNNING','SUCCESS','FAILED','CANCELLED') DEFAULT 'PENDING' COMMENT '执行状态',
--   result JSON DEFAULT NULL COMMENT '执行结果',
--   err_msg TEXT DEFAULT NULL COMMENT '错误信息',
--   FOREIGN KEY (task_def_id) REFERENCES task_def(id),
--   FOREIGN KEY (parent_instance_id) REFERENCES task_instance(id) ON DELETE CASCADE,
--   PRIMARY KEY (id),
--   INDEX idx_pii_s (parent_instance_id, ord)
-- ) COMMENT='任务实例表';
