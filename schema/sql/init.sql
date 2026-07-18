-- =============================================
-- 1. 命名空间表：顶层隔离维度
-- =============================================
CREATE TABLE `namespaces` (
  `id`          BIGINT UNSIGNED NOT NULL COMMENT '雪花算法主键ID',
  `namespace`   VARCHAR(64)  NOT NULL COMMENT '命名空间唯一标识',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '命名空间描述',
  `version`     INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '记录版本号，乐观锁',
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(毫秒)',
  `updated_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(毫秒)',
  `deleted_at`  DATETIME(3) NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_namespace` (`namespace`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='命名空间表';

-- =============================================
-- 2. 配置表：每个 config_name 对应一份完整配置文件
-- =============================================
CREATE TABLE `configs` (
  `id`           BIGINT UNSIGNED NOT NULL COMMENT '雪花算法主键ID',
  `namespace_id` BIGINT UNSIGNED NOT NULL COMMENT '关联命名空间ID',
  `config_name`  VARCHAR(255) NOT NULL COMMENT '配置名称，如 app / db',
  `format`       TINYINT      NOT NULL DEFAULT 1 COMMENT '配置格式：1=json, 2=yaml, 3=text, 4=properties',
  `content`      LONGTEXT     NOT NULL COMMENT '配置内容全文，支持任意格式文本',
  `description`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置说明',
  `created_by`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人（Casdoor用户名）',
  `updated_by`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '最后更新人',
  `version`      INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '配置内容版本号，每次更新自增',
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(毫秒)',
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(毫秒)',
  `deleted_at`   DATETIME(3) NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_namespace_config` (`namespace_id`, `config_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置主表';

-- =============================================
-- 3. 配置历史表：全量快照留存，支持版本回滚与审计
-- =============================================
CREATE TABLE `config_history` (
  `id`            BIGINT UNSIGNED NOT NULL COMMENT '雪花算法主键ID',
  `config_id`     BIGINT UNSIGNED NOT NULL COMMENT '关联配置ID',
  `namespace_id`  BIGINT UNSIGNED NOT NULL COMMENT '冗余命名空间ID，加速查询',
  `config_name`   VARCHAR(255) NOT NULL COMMENT '冗余配置名，配置删除后仍可追溯',
  `content`       LONGTEXT     NOT NULL COMMENT '该版本完整配置快照',
  `config_version`       INT UNSIGNED NOT NULL COMMENT '对应配置内容版本号',
  `operator`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作人',
  `change_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变更原因',
  `version`      INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号',
  `created_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(毫秒)',
  `updated_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(毫秒)',
  `deleted_at`    DATETIME(3) NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置历史版本表';

-- =============================================
-- 4. 访问密钥表：SDK 端独立鉴权凭证
-- =============================================
CREATE TABLE `access_keys` (
  `id`          BIGINT UNSIGNED NOT NULL COMMENT '雪花算法主键ID',
  `access_key`  VARCHAR(64)  NOT NULL COMMENT '密钥ID（对外暴露）',
  `secret_hash` VARCHAR(128) NOT NULL COMMENT '密钥哈希值（bcrypt存储）',
  `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密钥用途说明',
  `version`     INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '记录乐观锁版本',
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(毫秒)',
  `updated_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(毫秒)',
  `deleted_at`  DATETIME(3) NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_access_key` (`access_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='访问密钥表';

-- =============================================
-- 5. 密钥-命名空间关联表：多对多挂载关系
-- =============================================
CREATE TABLE `access_key_namespaces` (
  `id`            BIGINT UNSIGNED NOT NULL COMMENT '雪花算法主键ID',
  `access_key_id` BIGINT UNSIGNED NOT NULL COMMENT '关联密钥ID',
  `namespace_id`  BIGINT UNSIGNED NOT NULL COMMENT '关联命名空间ID',
  `permission`    TINYINT      NOT NULL DEFAULT 1 COMMENT '权限：1=只读 / 2=读写',
  `version`       INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '记录乐观锁版本',
  `created_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(毫秒)',
  `updated_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(毫秒)',
  `deleted_at`    DATETIME(3) NULL COMMENT '软删除时间，NULL=未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key_namespace` (`access_key_id`, `namespace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密钥命名空间授权关联表';
