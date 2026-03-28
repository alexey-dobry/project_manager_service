-- groups + group_memberships
-- UUID генерирует приложение.

CREATE TABLE IF NOT EXISTS groups (
    id          UUID PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    course      INTEGER      NOT NULL CHECK (course > 0 AND course <= 10),
    faculty     VARCHAR(120) NOT NULL,
    leader_id   UUID         NOT NULL,                  -- FK в auth-service.users.id (без жёсткого FK, БД у каждого сервиса своя)
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_groups_faculty ON groups (faculty);
CREATE INDEX IF NOT EXISTS idx_groups_course  ON groups (course);
CREATE INDEX IF NOT EXISTS idx_groups_leader  ON groups (leader_id);

CREATE TABLE IF NOT EXISTS group_memberships (
    group_id      UUID         NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id       UUID         NOT NULL,                -- FK в auth — логический, не БД-уровня
    role_in_group VARCHAR(32)  NOT NULL CHECK (role_in_group IN ('member','leader')),
    joined_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_memberships_user ON group_memberships (user_id);
