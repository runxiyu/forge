-- SPDX-License-Identifier: AGPL-3.0-only
-- SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

-- Currently, slugs accept arbitrary unicode text. We should
-- look into normalization options later.
-- May consider using citext and limiting it to safe characters.

CREATE TABLE groups (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT NOT NULL,
	parent_group BIGINT REFERENCES groups(id) ON DELETE RESTRICT,
	description TEXT,
	UNIQUE NULLS NOT DISTINCT (parent_group, name)
);
CREATE INDEX IF NOT EXISTS groups_parent_idx ON groups(parent_group);

DO $$ BEGIN
	CREATE TYPE contrib_requirement AS ENUM ('closed','registered_user','federated','ssh_pubkey','open');
	-- closed means only those with direct access; each layer adds that level of user
EXCEPTION WHEN duplicate_object THEN END $$;
CREATE TABLE repos (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT, -- I mean, should be CASCADE but deleting Git repos on disk also needs to be considered
	name TEXT NOT NULL,
	description TEXT,
	contrib_requirements contrib_requirement NOT NULL,
	filesystem_path TEXT NOT NULL, -- does not have to be unique, double-mounting is allowed
	UNIQUE(group_id, name)
);
CREATE INDEX IF NOT EXISTS repos_group_idx ON repos(group_id);

CREATE TABLE mailing_lists (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	description TEXT,
	UNIQUE(group_id, name)
);
CREATE INDEX IF NOT EXISTS mailing_lists_group_idx ON mailing_lists(group_id);

CREATE TABLE mailing_list_emails (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	list_id BIGINT NOT NULL REFERENCES mailing_lists(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	sender TEXT NOT NULL,
	date TIMESTAMPTZ NOT NULL, -- everything must be in UTC
	message_id TEXT, -- no uniqueness guarantee as it's arbitrarily set by senders
	content BYTEA NOT NULL
);

DO $$ BEGIN
	CREATE TYPE user_type AS ENUM ('pubkey_only','federated','registered','admin');
EXCEPTION WHEN duplicate_object THEN END $$;
CREATE TABLE users (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username TEXT UNIQUE, -- NULL when, for example, pubkey_only
	type user_type NOT NULL,
	password_hash TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ssh_public_keys (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	key_string TEXT NOT NULL,
	CONSTRAINT unique_key_string EXCLUDE USING HASH (key_string WITH =) -- because apparently some haxxor like using rsa16384 keys which are too long for a simple UNIQUE constraint :D
);
CREATE INDEX IF NOT EXISTS ssh_keys_user_idx ON ssh_public_keys(user_id);

CREATE TABLE sessions (
	session_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash BYTEA UNIQUE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_user_idx   ON sessions(user_id);

DO $$ BEGIN
	CREATE TYPE group_role AS ENUM ('owner'); -- just owner for now, might need to rethink ACL altogether later; might consider using a join table if we need it to be dynamic, but enum suffices for now
EXCEPTION WHEN duplicate_object THEN END $$;
CREATE TABLE user_group_roles (
	group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role group_role NOT NULL,
	PRIMARY KEY(user_id, group_id)
);
CREATE INDEX IF NOT EXISTS ugr_group_idx ON user_group_roles(group_id);

CREATE TABLE federated_identities (
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	service TEXT NOT NULL, -- might need to constrain
	remote_username TEXT NOT NULL,
	PRIMARY KEY(user_id, service),
	UNIQUE(service, remote_username)
);

CREATE TABLE ticket_trackers (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	description TEXT,
	UNIQUE(group_id, name)
);

CREATE TABLE tickets (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tracker_id BIGINT NOT NULL REFERENCES ticket_trackers(id) ON DELETE CASCADE,
	tracker_local_id BIGINT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	UNIQUE(tracker_id, tracker_local_id)
);

CREATE OR REPLACE FUNCTION create_tracker_ticket_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('tracker_ticket_seq_%s', NEW.id);
BEGIN
	EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %I', seq_name);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION drop_tracker_ticket_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('tracker_ticket_seq_%s', OLD.id);
BEGIN
	EXECUTE format('DROP SEQUENCE IF EXISTS %I', seq_name);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS after_insert_ticket_tracker ON ticket_trackers;
CREATE TRIGGER after_insert_ticket_tracker
AFTER INSERT ON ticket_trackers
FOR EACH ROW
EXECUTE FUNCTION create_tracker_ticket_sequence();
DROP TRIGGER IF EXISTS before_delete_ticket_tracker ON ticket_trackers;
CREATE TRIGGER before_delete_ticket_tracker
BEFORE DELETE ON ticket_trackers
FOR EACH ROW
EXECUTE FUNCTION drop_tracker_ticket_sequence();
CREATE OR REPLACE FUNCTION assign_tracker_local_id()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('tracker_ticket_seq_%s', NEW.tracker_id);
BEGIN
	IF NEW.tracker_local_id IS NULL THEN
		EXECUTE format('SELECT nextval(%L)', seq_name) INTO NEW.tracker_local_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS before_insert_ticket ON tickets;
CREATE TRIGGER before_insert_ticket
BEFORE INSERT ON tickets
FOR EACH ROW
EXECUTE FUNCTION assign_tracker_local_id();
CREATE INDEX IF NOT EXISTS tickets_tracker_idx ON tickets(tracker_id);

DO $$ BEGIN
	CREATE TYPE mr_status AS ENUM ('open','merged','closed');
EXCEPTION WHEN duplicate_object THEN END $$;

CREATE TABLE merge_requests (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	repo_id BIGINT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
	repo_local_id BIGINT NOT NULL,
	title TEXT NOT NULL,
	creator BIGINT REFERENCES users(id) ON DELETE SET NULL,
	source_repo BIGINT NOT NULL REFERENCES repos(id) ON DELETE RESTRICT,
	source_ref TEXT NOT NULL,
	destination_branch TEXT,
	status mr_status NOT NULL,
	UNIQUE (repo_id, repo_local_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS mr_open_src_dst_uniq
	ON merge_requests (repo_id, source_repo, source_ref, coalesce(destination_branch, ''))
	WHERE status = 'open';
CREATE INDEX IF NOT EXISTS mr_repo_idx    ON merge_requests(repo_id);
CREATE INDEX IF NOT EXISTS mr_creator_idx ON merge_requests(creator);
CREATE OR REPLACE FUNCTION create_repo_mr_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('repo_mr_seq_%s', NEW.id);
BEGIN
	EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %I', seq_name);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION drop_repo_mr_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('repo_mr_seq_%s', OLD.id);
BEGIN
	EXECUTE format('DROP SEQUENCE IF EXISTS %I', seq_name);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS after_insert_repo ON repos;
CREATE TRIGGER after_insert_repo
AFTER INSERT ON repos
FOR EACH ROW
EXECUTE FUNCTION create_repo_mr_sequence();
DROP TRIGGER IF EXISTS before_delete_repo ON repos;
CREATE TRIGGER before_delete_repo
BEFORE DELETE ON repos
FOR EACH ROW
EXECUTE FUNCTION drop_repo_mr_sequence();
CREATE OR REPLACE FUNCTION assign_repo_local_id()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := format('repo_mr_seq_%s', NEW.repo_id);
BEGIN
	IF NEW.repo_local_id IS NULL THEN
		EXECUTE format('SELECT nextval(%L)', seq_name) INTO NEW.repo_local_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS before_insert_merge_request ON merge_requests;
CREATE TRIGGER before_insert_merge_request
BEFORE INSERT ON merge_requests
FOR EACH ROW
EXECUTE FUNCTION assign_repo_local_id();
