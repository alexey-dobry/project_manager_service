-- projects + tasks + comments
-- UUID генерирует приложение.

CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    group_id    UUID         NOT NULL,                       -- логическая FK -> groups-service
    owner_id    UUID         NOT NULL,                       -- логическая FK -> auth-service.users
    status      VARCHAR(32)  NOT NULL CHECK (status IN ('draft','in_progress','review','completed','archived')),
    deadline    TIMESTAMPTZ  NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_group_id ON projects (group_id);
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects (owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_status   ON projects (status);

CREATE TABLE IF NOT EXISTS tasks (
    id          UUID PRIMARY KEY,
    project_id  UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title       VARCHAR(200) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    assignee_id UUID         NULL,                          -- логическая FK
    status      VARCHAR(32)  NOT NULL CHECK (status IN ('todo','in_progress','done','blocked')),
    priority    VARCHAR(32)  NOT NULL CHECK (priority IN ('low','medium','high','critical')),
    due_date    TIMESTAMPTZ  NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_project_id  ON tasks (project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status      ON tasks (status);
-- Композитный индекс под /stats и фильтры по статусу внутри проекта.
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks (project_id, status);

CREATE TABLE IF NOT EXISTS comments (
    id         UUID PRIMARY KEY,
    task_id    UUID         NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL,                       -- логическая FK
    content    TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_comments_task_id ON comments (task_id);
