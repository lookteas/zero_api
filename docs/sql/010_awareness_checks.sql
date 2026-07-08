CREATE TABLE IF NOT EXISTS awareness_checks (
  check_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '意识强度检测ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  status VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT '状态：draft=草稿，in_progress=检测中，completed=已完成，abandoned=已放弃',
  done_chapters INT NOT NULL DEFAULT 0 COMMENT '已完成章节数',
  total_chapters INT NOT NULL DEFAULT 9 COMMENT '总章节数',
  score DECIMAL(6,2) DEFAULT NULL COMMENT '综合得分',
  ref_score DECIMAL(6,2) DEFAULT NULL COMMENT '综合人类平均参考分',
  delta DECIMAL(6,2) DEFAULT NULL COMMENT '综合分相对人类平均参考差值',
  prev_check_id BIGINT UNSIGNED DEFAULT NULL COMMENT '上一轮可对比检测ID',
  started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '检测开始时间',
  completed_at TIMESTAMP NULL DEFAULT NULL COMMENT '检测完成时间',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (check_id),
  KEY idx_awareness_checks_user_status (user_id, status, check_id),
  KEY idx_awareness_checks_user_completed (user_id, completed_at, check_id),
  KEY idx_awareness_checks_prev (prev_check_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识强度检测轮次表';

CREATE TABLE IF NOT EXISTS awareness_check_chapters (
  check_chapter_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '意识强度检测章节ID',
  check_id BIGINT UNSIGNED NOT NULL COMMENT '检测ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节ID',
  total_points INT NOT NULL DEFAULT 0 COMMENT '本章可检测意识点数量',
  scored_points INT NOT NULL DEFAULT 0 COMMENT '已评分点数',
  score DECIMAL(6,2) DEFAULT NULL COMMENT '章节得分',
  ref_score DECIMAL(6,2) DEFAULT NULL COMMENT '章节人类平均参考分',
  delta DECIMAL(6,2) DEFAULT NULL COMMENT '章节分相对人类平均参考差值',
  prev_score DECIMAL(6,2) DEFAULT NULL COMMENT '上一轮同章节得分快照',
  score_change DECIMAL(6,2) DEFAULT NULL COMMENT '本章相对上一轮变化',
  status VARCHAR(20) NOT NULL DEFAULT 'not_started' COMMENT '状态：not_started=未检测，in_progress=检测中，completed=已完成',
  submitted_at TIMESTAMP NULL DEFAULT NULL COMMENT '章节提交时间',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (check_chapter_id),
  UNIQUE KEY uk_awareness_check_chapter (check_id, chapter_id),
  KEY idx_awareness_check_chapters_user (user_id, check_id),
  KEY idx_awareness_check_chapters_status (check_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识强度检测章节状态表';

CREATE TABLE IF NOT EXISTS awareness_check_scores (
  score_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '意识强度检测评分ID',
  check_id BIGINT UNSIGNED NOT NULL COMMENT '检测ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节ID',
  awareness_id BIGINT UNSIGNED NOT NULL COMMENT '意识强度点ID',
  self_score DECIMAL(5,2) NOT NULL DEFAULT 50.00 COMMENT '原始自评分，百分比0.00-100.00',
  score DECIMAL(6,2) NOT NULL DEFAULT 50.00 COMMENT '按方向换算后的得分',
  ref_score DECIMAL(6,2) NOT NULL DEFAULT 50.00 COMMENT '按方向换算后的人类平均参考分',
  delta DECIMAL(6,2) NOT NULL DEFAULT 0.00 COMMENT '得分相对人类平均参考差值',
  prev_score DECIMAL(6,2) DEFAULT NULL COMMENT '上一轮同点位得分快照',
  score_change DECIMAL(6,2) DEFAULT NULL COMMENT '本点位相对上一轮变化',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (score_id),
  UNIQUE KEY uk_awareness_check_score (check_id, awareness_id),
  KEY idx_awareness_check_scores_user (user_id, check_id),
  KEY idx_awareness_check_scores_chapter (check_id, chapter_id),
  KEY idx_awareness_check_scores_awareness (awareness_id),
  CHECK (self_score >= 0 AND self_score <= 100),
  CHECK (score >= 0 AND score <= 100),
  CHECK (ref_score >= 0 AND ref_score <= 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识强度检测评分表';

SET @check_chapter_no_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_chapters'
    AND column_name = 'chapter_no'
);
SET @check_chapter_no_sql = IF(
  @check_chapter_no_exists > 0,
  'ALTER TABLE awareness_check_chapters DROP COLUMN chapter_no',
  'SELECT 1'
);
PREPARE check_chapter_no_stmt FROM @check_chapter_no_sql;
EXECUTE check_chapter_no_stmt;
DEALLOCATE PREPARE check_chapter_no_stmt;

SET @check_chapter_title_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_chapters'
    AND column_name = 'chapter_title'
);
SET @check_chapter_title_sql = IF(
  @check_chapter_title_exists > 0,
  'ALTER TABLE awareness_check_chapters DROP COLUMN chapter_title',
  'SELECT 1'
);
PREPARE check_chapter_title_stmt FROM @check_chapter_title_sql;
EXECUTE check_chapter_title_stmt;
DEALLOCATE PREPARE check_chapter_title_stmt;

SET @check_chapter_full_title_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_chapters'
    AND column_name = 'chapter_full_title'
);
SET @check_chapter_full_title_sql = IF(
  @check_chapter_full_title_exists > 0,
  'ALTER TABLE awareness_check_chapters DROP COLUMN chapter_full_title',
  'SELECT 1'
);
PREPARE check_chapter_full_title_stmt FROM @check_chapter_full_title_sql;
EXECUTE check_chapter_full_title_stmt;
DEALLOCATE PREPARE check_chapter_full_title_stmt;

SET @check_score_title_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_scores'
    AND column_name = 'title'
);
SET @check_score_title_sql = IF(
  @check_score_title_exists > 0,
  'ALTER TABLE awareness_check_scores DROP COLUMN title',
  'SELECT 1'
);
PREPARE check_score_title_stmt FROM @check_score_title_sql;
EXECUTE check_score_title_stmt;
DEALLOCATE PREPARE check_score_title_stmt;

SET @check_score_summary_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_scores'
    AND column_name = 'summary'
);
SET @check_score_summary_sql = IF(
  @check_score_summary_exists > 0,
  'ALTER TABLE awareness_check_scores DROP COLUMN summary',
  'SELECT 1'
);
PREPARE check_score_summary_stmt FROM @check_score_summary_sql;
EXECUTE check_score_summary_stmt;
DEALLOCATE PREPARE check_score_summary_stmt;

SET @check_score_human_score_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_scores'
    AND column_name = 'human_score'
);
SET @check_score_human_score_sql = IF(
  @check_score_human_score_exists > 0,
  'ALTER TABLE awareness_check_scores DROP COLUMN human_score',
  'SELECT 1'
);
PREPARE check_score_human_score_stmt FROM @check_score_human_score_sql;
EXECUTE check_score_human_score_stmt;
DEALLOCATE PREPARE check_score_human_score_stmt;

SET @check_score_direction_exists = (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'awareness_check_scores'
    AND column_name = 'direction'
);
SET @check_score_direction_sql = IF(
  @check_score_direction_exists > 0,
  'ALTER TABLE awareness_check_scores DROP COLUMN direction',
  'SELECT 1'
);
PREPARE check_score_direction_stmt FROM @check_score_direction_sql;
EXECUTE check_score_direction_stmt;
DEALLOCATE PREPARE check_score_direction_stmt;
