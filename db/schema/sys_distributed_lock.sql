CREATE TABLE IF NOT EXISTS sys_distributed_lock (
    id BIGINT AUTO_INCREMENT NOT NULL,
    gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    lock_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '锁的唯一标识',
    lock_holder VARCHAR(255) NOT NULL DEFAULT '' COMMENT '锁持有者标识，用于验证所有权',
    lock_ttl INT NOT NULL DEFAULT 0 COMMENT '锁的存活时间（秒），0表示不过期',
    expire_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '锁过期时间',
    is_active TINYINT NOT NULL DEFAULT 1 COMMENT '锁是否激活：1=激活，0=已释放或过期',
    PRIMARY KEY (id),
    UNIQUE KEY uk_lock_key (lock_key),
    KEY idx_expire_time (expire_time),
    KEY idx_lock_holder (lock_holder)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';
