DROP INDEX IF EXISTS idx_tasks_project_status_position;
ALTER TABLE tasks DROP COLUMN IF EXISTS position;
