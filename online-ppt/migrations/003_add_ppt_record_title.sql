-- 003_add_ppt_record_title.sql
-- Adds title field to ppt_records table for user-friendly display names.

ALTER TABLE ppt_records 
ADD COLUMN title VARCHAR(255) NULL AFTER name;
