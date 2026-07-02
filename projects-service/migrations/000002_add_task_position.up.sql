-- position — порядок карточки задачи внутри колонки Kanban-доски (сортировка
-- внутри одного статуса). Название "order" зарезервировано в SQL, поэтому
-- столбец называется position.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS position INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_tasks_project_status_position ON tasks (project_id, status, position);
