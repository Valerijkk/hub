-- TASK-098 / P1-BE-U03 — news_poll* (FR-INTR-01..03, 07/08)
-- Задача: .planning/backlog/TASK-098-2026-07-28-intranet-author-delete-polls-adr.md
-- ADR: .planning/decisions/ADR-INTRANET-U03-NEWS-POLL-AUTHOR-DELETE-2026.md
--
-- Hub в test/dev использует Ent Schema.Create; этот файл — канон для
-- существующих Postgres-инстансов. Owner-approve migrate не требуется (2026-07-28).
-- Soft-delete 30d поста сохраняется; poll-строки без собственного deleted_at.

CREATE TABLE IF NOT EXISTS news_polls (
	id UUID PRIMARY KEY,
	news_post_id UUID NOT NULL UNIQUE REFERENCES news_posts(id) ON DELETE CASCADE,
	question VARCHAR(500) NOT NULL,
	closed BOOLEAN NOT NULL DEFAULT FALSE,
	version BIGINT NOT NULL DEFAULT 1,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS news_poll_options (
	id UUID PRIMARY KEY,
	poll_id UUID NOT NULL REFERENCES news_polls(id) ON DELETE CASCADE,
	text VARCHAR(200) NOT NULL,
	position BIGINT NOT NULL DEFAULT 0,
	votes_count BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS newspolloption_poll_id_position
	ON news_poll_options (poll_id, position);

CREATE TABLE IF NOT EXISTS news_poll_votes (
	id UUID PRIMARY KEY,
	poll_id UUID NOT NULL REFERENCES news_polls(id) ON DELETE CASCADE,
	option_id UUID NOT NULL REFERENCES news_poll_options(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS newspollvote_poll_id_user_id
	ON news_poll_votes (poll_id, user_id);

CREATE INDEX IF NOT EXISTS newspollvote_option_id
	ON news_poll_votes (option_id);
