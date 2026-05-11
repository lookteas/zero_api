CREATE TABLE IF NOT EXISTS app_settings (
  setting_key VARCHAR(100) NOT NULL,
  setting_value VARCHAR(500) NOT NULL DEFAULT '',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用设置表';

INSERT INTO app_settings (setting_key, setting_value)
VALUES
  ('awareness_cycle_start_date', '2026-05-01'),
  ('awareness_cycle_rest_days', '7')
ON DUPLICATE KEY UPDATE setting_value = setting_value;

SET @daily_task_awareness_id_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND column_name = 'awareness_id'
);

SET @daily_task_awareness_id_sql = IF(
  @daily_task_awareness_id_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD COLUMN awareness_id BIGINT UNSIGNED DEFAULT NULL AFTER topic_id'
);
PREPARE daily_task_awareness_id_stmt FROM @daily_task_awareness_id_sql;
EXECUTE daily_task_awareness_id_stmt;
DEALLOCATE PREPARE daily_task_awareness_id_stmt;

SET @daily_task_awareness_idx_exists = (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'daily_tasks'
    AND index_name = 'idx_awareness_id'
);

SET @daily_task_awareness_idx_sql = IF(
  @daily_task_awareness_idx_exists > 0,
  'SELECT 1',
  'ALTER TABLE daily_tasks ADD KEY idx_awareness_id (awareness_id)'
);
PREPARE daily_task_awareness_idx_stmt FROM @daily_task_awareness_idx_sql;
EXECUTE daily_task_awareness_idx_stmt;
DEALLOCATE PREPARE daily_task_awareness_idx_stmt;

UPDATE daily_tasks dt
JOIN (
  SELECT point_title, MIN(awareness_id) AS awareness_id
  FROM awareness
  WHERE status = 1
    AND is_meta = 0
  GROUP BY point_title
  HAVING COUNT(*) = 1
) matched_awareness ON matched_awareness.point_title = dt.topic_title
SET dt.awareness_id = matched_awareness.awareness_id
WHERE dt.awareness_id IS NULL;
