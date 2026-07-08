CREATE TABLE IF NOT EXISTS communities (
  community_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '社群ID，SaaS模式下用于隔离不同社群或空间的数据',
  community_name VARCHAR(100) NOT NULL COMMENT '社群名称，例如默认社群、训练营名称',
  community_code VARCHAR(64) DEFAULT NULL COMMENT '社群唯一编码，可用于邀请链接、二级标识或外部系统映射',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用，0=停用',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (community_id),
  UNIQUE KEY uk_communities_code (community_code),
  KEY idx_communities_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='社群表';

INSERT INTO communities (community_id, community_name, community_code, status)
VALUES (1, '默认社群', 'default', 1)
ON DUPLICATE KEY UPDATE
  community_name = community_name,
  community_code = community_code,
  status = status;

CREATE TABLE IF NOT EXISTS awareness_cycles (
  cycle_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '意识打卡轮次ID',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '所属社群ID；当前个人站点默认使用1，SaaS模式下每个社群可有自己的轮次',
  cycle_name VARCHAR(100) NOT NULL DEFAULT '默认意识打卡轮次' COMMENT '轮次名称，用于后台识别，例如2026春季意识提升营',
  start_date DATE NOT NULL COMMENT '轮次启动日期；从该日期开始计算第一个有效打卡日',
  rest_days INT NOT NULL DEFAULT 7 COMMENT '每轮所有意识点完成后的固定休息天数',
  schedule_horizon_days INT NOT NULL DEFAULT 365 COMMENT '每次重算时生成未来多少天的排程',
  status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '轮次状态：draft=草稿，active=启用，paused=整体暂停，archived=归档',
  last_generated_until DATE DEFAULT NULL COMMENT '排程已生成到的最后日期',
  created_by BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人用户ID或管理员ID，当前可为空',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (cycle_id),
  KEY idx_awareness_cycles_community_status (community_id, status),
  KEY idx_awareness_cycles_start_date (start_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识打卡轮次配置表';

INSERT INTO awareness_cycles (cycle_id, community_id, cycle_name, start_date, rest_days, schedule_horizon_days, status)
SELECT
  1,
  1,
  '默认意识打卡轮次',
  COALESCE(
    STR_TO_DATE((SELECT setting_value FROM app_settings WHERE setting_key = 'awareness_cycle_start_date' LIMIT 1), '%Y-%m-%d'),
    DATE('2026-05-01')
  ),
  COALESCE(
    CAST((SELECT setting_value FROM app_settings WHERE setting_key = 'awareness_cycle_rest_days' LIMIT 1) AS UNSIGNED),
    7
  ),
  365,
  'active'
WHERE NOT EXISTS (
  SELECT 1 FROM awareness_cycles WHERE cycle_id = 1
);

CREATE TABLE IF NOT EXISTS awareness_cycle_pauses (
  pause_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '暂停配置ID',
  cycle_id BIGINT UNSIGNED NOT NULL COMMENT '所属意识打卡轮次ID',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '所属社群ID，冗余存储便于按社群查询和隔离',
  pause_start_date DATE NOT NULL COMMENT '暂停开始日期，包含当天',
  pause_end_date DATE NOT NULL COMMENT '暂停结束日期，包含当天；单日暂停时等于开始日期',
  reason VARCHAR(255) DEFAULT NULL COMMENT '暂停原因，例如五一、春节、社群活动调整',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=生效，0=停用',
  created_by BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人用户ID或管理员ID，当前可为空',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (pause_id),
  KEY idx_cycle_pauses_cycle_date (cycle_id, pause_start_date, pause_end_date),
  KEY idx_cycle_pauses_community_date (community_id, pause_start_date, pause_end_date),
  KEY idx_cycle_pauses_status (status),
  CHECK (pause_end_date >= pause_start_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识打卡轮次暂停日期表';

CREATE TABLE IF NOT EXISTS awareness_schedule_days (
  schedule_day_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '每日排程ID',
  cycle_id BIGINT UNSIGNED NOT NULL COMMENT '所属意识打卡轮次ID',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '所属社群ID，用于SaaS多社群隔离',
  schedule_date DATE NOT NULL COMMENT '排程日期',
  day_type VARCHAR(20) NOT NULL COMMENT '日期类型：normal=正常打卡，paused=暂停打卡，rest=轮次结束后的固定休息日',
  awareness_id BIGINT UNSIGNED DEFAULT NULL COMMENT '意识强度点ID；normal时有值，paused/rest时为空',
  cycle_index INT NOT NULL DEFAULT 0 COMMENT '第几轮，从0开始计数',
  cycle_day_index INT DEFAULT NULL COMMENT '当前轮中的有效打卡日序号，从0开始；paused/rest为空',
  effective_day_index INT DEFAULT NULL COMMENT '从启动日开始累计的有效打卡日序号，从0开始；暂停日不计入',
  pause_id BIGINT UNSIGNED DEFAULT NULL COMMENT '对应暂停配置ID；day_type=paused时可有值',
  pause_reason VARCHAR(255) DEFAULT NULL COMMENT '暂停原因快照，例如五一、春节',
  generated_version BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '排程生成版本号；每次重算可递增，便于排查排程来源',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (schedule_day_id),
  UNIQUE KEY uk_schedule_cycle_date (cycle_id, schedule_date),
  KEY idx_schedule_community_date (community_id, schedule_date),
  KEY idx_schedule_day_type (day_type),
  KEY idx_schedule_awareness_id (awareness_id),
  KEY idx_schedule_pause_id (pause_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识打卡每日排程表';

SET @schedule_awareness_title_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'awareness_title'
);
SET @schedule_awareness_title_sql = IF(
  @schedule_awareness_title_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN awareness_title',
  'SELECT 1'
);
PREPARE schedule_awareness_title_stmt FROM @schedule_awareness_title_sql;
EXECUTE schedule_awareness_title_stmt;
DEALLOCATE PREPARE schedule_awareness_title_stmt;

SET @schedule_awareness_theme_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'awareness_theme'
);
SET @schedule_awareness_theme_sql = IF(
  @schedule_awareness_theme_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN awareness_theme',
  'SELECT 1'
);
PREPARE schedule_awareness_theme_stmt FROM @schedule_awareness_theme_sql;
EXECUTE schedule_awareness_theme_stmt;
DEALLOCATE PREPARE schedule_awareness_theme_stmt;

SET @schedule_awareness_summary_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'awareness_summary'
);
SET @schedule_awareness_summary_sql = IF(
  @schedule_awareness_summary_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN awareness_summary',
  'SELECT 1'
);
PREPARE schedule_awareness_summary_stmt FROM @schedule_awareness_summary_sql;
EXECUTE schedule_awareness_summary_stmt;
DEALLOCATE PREPARE schedule_awareness_summary_stmt;

SET @schedule_awareness_details_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'awareness_details'
);
SET @schedule_awareness_details_sql = IF(
  @schedule_awareness_details_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN awareness_details',
  'SELECT 1'
);
PREPARE schedule_awareness_details_stmt FROM @schedule_awareness_details_sql;
EXECUTE schedule_awareness_details_stmt;
DEALLOCATE PREPARE schedule_awareness_details_stmt;

SET @schedule_reference_min_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'reference_min'
);
SET @schedule_reference_min_sql = IF(
  @schedule_reference_min_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN reference_min',
  'SELECT 1'
);
PREPARE schedule_reference_min_stmt FROM @schedule_reference_min_sql;
EXECUTE schedule_reference_min_stmt;
DEALLOCATE PREPARE schedule_reference_min_stmt;

SET @schedule_reference_max_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'reference_max'
);
SET @schedule_reference_max_sql = IF(
  @schedule_reference_max_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN reference_max',
  'SELECT 1'
);
PREPARE schedule_reference_max_stmt FROM @schedule_reference_max_sql;
EXECUTE schedule_reference_max_stmt;
DEALLOCATE PREPARE schedule_reference_max_stmt;

SET @schedule_better_direction_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_schedule_days'
    AND column_name = 'better_direction'
);
SET @schedule_better_direction_sql = IF(
  @schedule_better_direction_exists > 0,
  'ALTER TABLE awareness_schedule_days DROP COLUMN better_direction',
  'SELECT 1'
);
PREPARE schedule_better_direction_stmt FROM @schedule_better_direction_sql;
EXECUTE schedule_better_direction_stmt;
DEALLOCATE PREPARE schedule_better_direction_stmt;

SET @daily_tasks_community_id_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND column_name = 'community_id'
);

SET @daily_tasks_community_id_sql = IF(
  @daily_tasks_community_id_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD COLUMN community_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT ''所属社群ID；当前个人站点默认1，SaaS模式下用于区分不同社群打卡'' AFTER user_id'
);
PREPARE daily_tasks_community_id_stmt FROM @daily_tasks_community_id_sql;
EXECUTE daily_tasks_community_id_stmt;
DEALLOCATE PREPARE daily_tasks_community_id_stmt;

SET @daily_tasks_schedule_day_id_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND column_name = 'schedule_day_id'
);

SET @daily_tasks_schedule_day_id_sql = IF(
  @daily_tasks_schedule_day_id_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD COLUMN schedule_day_id BIGINT UNSIGNED DEFAULT NULL COMMENT ''关联的每日排程ID；用于追溯该打卡来自哪一天的排程'' AFTER task_date'
);
PREPARE daily_tasks_schedule_day_id_stmt FROM @daily_tasks_schedule_day_id_sql;
EXECUTE daily_tasks_schedule_day_id_stmt;
DEALLOCATE PREPARE daily_tasks_schedule_day_id_stmt;

SET @daily_tasks_community_date_idx_exists = (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND index_name = 'idx_daily_tasks_community_date'
);

SET @daily_tasks_community_date_idx_sql = IF(
  @daily_tasks_community_date_idx_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD KEY idx_daily_tasks_community_date (community_id, task_date)'
);
PREPARE daily_tasks_community_date_idx_stmt FROM @daily_tasks_community_date_idx_sql;
EXECUTE daily_tasks_community_date_idx_stmt;
DEALLOCATE PREPARE daily_tasks_community_date_idx_stmt;

SET @daily_tasks_schedule_day_id_idx_exists = (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND index_name = 'idx_daily_tasks_schedule_day_id'
);

SET @daily_tasks_schedule_day_id_idx_sql = IF(
  @daily_tasks_schedule_day_id_idx_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD KEY idx_daily_tasks_schedule_day_id (schedule_day_id)'
);
PREPARE daily_tasks_schedule_day_id_idx_stmt FROM @daily_tasks_schedule_day_id_idx_sql;
EXECUTE daily_tasks_schedule_day_id_idx_stmt;
DEALLOCATE PREPARE daily_tasks_schedule_day_id_idx_stmt;
