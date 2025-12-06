-- 初始化数据库脚本
-- 如果数据库不存在则创建
CREATE DATABASE IF NOT EXISTS forgetting_curve CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE forgetting_curve;

-- 注意：表结构会由 GORM 自动迁移创建
-- 这里可以添加一些初始化数据或索引优化

-- 创建索引优化查询性能（GORM 会自动创建，这里只是示例）
-- CREATE INDEX IF NOT EXISTS idx_students_open_id ON students(open_id);
-- CREATE INDEX IF NOT EXISTS idx_words_student_id ON words(student_id);
-- CREATE INDEX IF NOT EXISTS idx_words_start_date ON words(start_date);

