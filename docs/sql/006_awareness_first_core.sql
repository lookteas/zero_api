CREATE TABLE IF NOT EXISTS chapters (
  chapter_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '章节ID',
  chapter_no INT NOT NULL COMMENT '章节序号',
  chapter_area VARCHAR(100) DEFAULT NULL COMMENT '章节所属区域',
  chapter_title VARCHAR(255) NOT NULL COMMENT '章节标题',
  chapter_full_title VARCHAR(255) NOT NULL COMMENT '章节完整标题',
  sort_order INT NOT NULL DEFAULT 0 COMMENT '章节排序号',
  source_volume VARCHAR(20) DEFAULT NULL COMMENT '来源册别：上册/下册',
  notes TEXT DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (chapter_id),
  KEY idx_chapters_no (chapter_no),
  KEY idx_chapters_sort (sort_order),
  KEY idx_chapters_volume (source_volume)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节公共表';

CREATE TABLE IF NOT EXISTS app_settings (
  setting_key VARCHAR(100) NOT NULL,
  setting_value VARCHAR(500) NOT NULL DEFAULT '',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用设置表';

CREATE TABLE IF NOT EXISTS regions (
  region_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '区域ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节ID',
  region_title VARCHAR(255) NOT NULL COMMENT '区域标题',
  region_full_title VARCHAR(255) NOT NULL COMMENT '区域完整标题',
  sort_order INT NOT NULL DEFAULT 0 COMMENT '区域排序号',
  notes TEXT DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (region_id),
  KEY idx_regions_chapter_id (chapter_id),
  KEY idx_regions_sort (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='所属区域公共表';

CREATE TABLE IF NOT EXISTS sections (
  section_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '小节ID',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID',
  section_title VARCHAR(255) NOT NULL COMMENT '小节标题',
  section_full_title VARCHAR(255) NOT NULL COMMENT '小节完整标题',
  sort_order INT NOT NULL DEFAULT 0 COMMENT '小节排序号',
  notes TEXT DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (section_id),
  KEY idx_sections_region_id (region_id),
  KEY idx_sections_sort (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='小节公共表';

CREATE TABLE IF NOT EXISTS awareness (
  awareness_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '意识强度点ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节ID',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID',
  section_id BIGINT UNSIGNED NOT NULL COMMENT '小节ID',
  source_volume VARCHAR(20) DEFAULT NULL COMMENT '来源册别：上册/下册',
  source_file VARCHAR(255) DEFAULT NULL COMMENT '来源文件名',
  point_no VARCHAR(50) DEFAULT NULL COMMENT '原文条目序号',
  point_title VARCHAR(255) NOT NULL COMMENT '意识强度点标题',
  theme VARCHAR(255) DEFAULT NULL COMMENT '主题',
  summary TEXT DEFAULT NULL COMMENT '摘要',
  quantitative_data TEXT DEFAULT NULL COMMENT '量化数据',
  reference_min DECIMAL(5,2) DEFAULT NULL COMMENT '原文参考最小值，百分比范围0.00-100.00',
  reference_max DECIMAL(5,2) DEFAULT NULL COMMENT '原文参考最大值，百分比范围0.00-100.00',
  value_unit VARCHAR(20) NOT NULL DEFAULT 'percent' COMMENT '数值单位：percent=百分比',
  better_direction VARCHAR(20) NOT NULL DEFAULT 'higher' COMMENT '优化方向：higher=越高越好，lower=越低越好',
  details LONGTEXT DEFAULT NULL COMMENT '详情正文',
  is_meta TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否说明性条目：1是，0否',
  status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态：1=正常可展示/可参与计算，0=停用或需人工复核',
  has_images TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否有图片',
  image_count INT NOT NULL DEFAULT 0 COMMENT '图片数量',
  cover_image_id BIGINT UNSIGNED DEFAULT NULL COMMENT '封面图片ID',
  image_notes TEXT DEFAULT NULL COMMENT '图片备注',
  images_json JSON DEFAULT NULL COMMENT '图片JSON数据',
  sort_order_global INT NOT NULL DEFAULT 0 COMMENT '全局排序号，按10递增便于补录',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (awareness_id),
  KEY idx_awareness_chapter_id (chapter_id),
  KEY idx_awareness_region_id (region_id),
  KEY idx_awareness_section_id (section_id),
  KEY idx_awareness_point_no (point_no),
  KEY idx_awareness_is_meta (is_meta),
  KEY idx_awareness_status (status),
  KEY idx_awareness_reference_range (reference_min, reference_max),
  KEY idx_awareness_sort_global (sort_order_global),
  KEY idx_awareness_cover_image_id (cover_image_id),
  CHECK (reference_min IS NULL OR (reference_min >= 0 AND reference_min <= 100)),
  CHECK (reference_max IS NULL OR (reference_max >= 0 AND reference_max <= 100)),
  CHECK (reference_min IS NULL OR reference_max IS NULL OR reference_min <= reference_max)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意识强度点主表';

CREATE TABLE IF NOT EXISTS user_awareness_scores (
  score_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户打分ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  awareness_id BIGINT UNSIGNED NOT NULL COMMENT '意识强度点ID',
  user_value DECIMAL(5,2) NOT NULL COMMENT '用户提交数值，百分比范围0.00-100.00',
  value_type VARCHAR(20) NOT NULL DEFAULT 'percent' COMMENT '数值类型：percent',
  compare_status VARCHAR(30) DEFAULT NULL COMMENT '比较结果：below_range/in_range/above_range',
  score_note VARCHAR(255) DEFAULT NULL COMMENT '系统生成的简短说明',
  submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
  PRIMARY KEY (score_id),
  KEY idx_user_awareness_scores_user_point (user_id, awareness_id),
  KEY idx_user_awareness_scores_awareness_id (awareness_id),
  KEY idx_user_awareness_scores_submitted_at (submitted_at),
  CHECK (user_value >= 0 AND user_value <= 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户意识强度点打分记录表';

CREATE TABLE IF NOT EXISTS point_images (
  image_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '图片ID',
  point_id BIGINT UNSIGNED NOT NULL COMMENT '意识强度点ID',
  image_path VARCHAR(500) DEFAULT NULL COMMENT '图片本地路径',
  image_url VARCHAR(500) DEFAULT NULL COMMENT '图片URL',
  image_title VARCHAR(255) DEFAULT NULL COMMENT '图片标题',
  caption TEXT DEFAULT NULL COMMENT '图片说明文字',
  alt_text VARCHAR(500) DEFAULT NULL COMMENT '图片替代文本',
  description TEXT DEFAULT NULL COMMENT '图片描述',
  sort_order INT NOT NULL DEFAULT 0 COMMENT '图片排序号',
  image_type VARCHAR(50) DEFAULT NULL COMMENT '图片类型',
  source VARCHAR(255) DEFAULT NULL COMMENT '图片来源',
  copyright_note TEXT DEFAULT NULL COMMENT '版权备注',
  status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态：1=正常可展示，0=停用',
  PRIMARY KEY (image_id),
  KEY idx_point_images_point_id (point_id),
  KEY idx_point_images_sort (sort_order),
  KEY idx_point_images_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片表';

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
