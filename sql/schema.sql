-- SPDX-License-Identifier: AGPL-3.0-only
-- SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

CREATE TABLE groups (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT NOT NULL,
	parent_group INTEGER REFERENCES groups(id) ON DELETE CASCADE,
	description TEXT,
	UNIQUE NULLS NOT DISTINCT (parent_group, name)
);

CREATE TABLE repos (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT, -- I mean, should be CASCADE but deleting Git repos on disk also needs to be considered
	contrib_requirements TEXT NOT NULL CHECK (contrib_requirements IN ('closed', 'registered_user', 'federated', 'ssh_pubkey', 'public')),
	name TEXT NOT NULL,
	UNIQUE(group_id, name),
	description TEXT,
	filesystem_path TEXT
);

CREATE TABLE mailing_lists (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	UNIQUE(group_id, name),
	description TEXT
);

CREATE TABLE mailing_list_emails (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	list_id INTEGER NOT NULL REFERENCES mailing_lists(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	sender TEXT NOT NULL,
	date TIMESTAMP NOT NULL,
	content BYTEA NOT NULL
);

CREATE TABLE mailing_list_subscribers (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	list_id INTEGER NOT NULL REFERENCES mailing_lists(id) ON DELETE CASCADE,
	email TEXT NOT NULL,
	UNIQUE (list_id, email)
);

CREATE TABLE users (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username TEXT UNIQUE,
	type TEXT NOT NULL CHECK (type IN ('pubkey_only', 'federated', 'registered', 'admin')),
	password TEXT
);

CREATE TABLE ssh_public_keys (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	key_string TEXT NOT NULL,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT unique_key_string EXCLUDE USING HASH (key_string WITH =)
);

CREATE TABLE sessions (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	session_id TEXT PRIMARY KEY NOT NULL,
	UNIQUE(user_id, session_id)
);

CREATE TABLE user_group_roles (
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	PRIMARY KEY(user_id, group_id)
);

CREATE TABLE federated_identities (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	service TEXT NOT NULL,
	remote_username TEXT NOT NULL,
	PRIMARY KEY(user_id, service)
);

-- Ticket tracking

CREATE TABLE ticket_trackers (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	description TEXT,
	UNIQUE(group_id, name)
);

CREATE TABLE tickets (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tracker_id INTEGER NOT NULL REFERENCES ticket_trackers(id) ON DELETE CASCADE,
	tracker_local_id INTEGER NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	UNIQUE(tracker_id, tracker_local_id)
);

CREATE OR REPLACE FUNCTION create_tracker_ticket_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'tracker_ticket_seq_' || NEW.id;
BEGIN
	EXECUTE format('CREATE SEQUENCE %I', seq_name);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER after_insert_ticket_tracker
AFTER INSERT ON ticket_trackers
FOR EACH ROW
EXECUTE FUNCTION create_tracker_ticket_sequence();

CREATE OR REPLACE FUNCTION drop_tracker_ticket_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'tracker_ticket_seq_' || OLD.id;
BEGIN
	EXECUTE format('DROP SEQUENCE IF EXISTS %I', seq_name);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER before_delete_ticket_tracker
BEFORE DELETE ON ticket_trackers
FOR EACH ROW
EXECUTE FUNCTION drop_tracker_ticket_sequence();

CREATE OR REPLACE FUNCTION assign_tracker_local_id()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'tracker_ticket_seq_' || NEW.tracker_id;
BEGIN
	IF NEW.tracker_local_id IS NULL THEN
		EXECUTE format('SELECT nextval(%L)', seq_name)
		INTO NEW.tracker_local_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER before_insert_ticket
BEFORE INSERT ON tickets
FOR EACH ROW
EXECUTE FUNCTION assign_tracker_local_id();

-- Merge requests

CREATE TABLE merge_requests (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	repo_id INTEGER NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
	repo_local_id INTEGER NOT NULL,
	title TEXT,
	creator INTEGER REFERENCES users(id) ON DELETE SET NULL,
	source_ref TEXT NOT NULL,
	destination_branch TEXT,
	status TEXT NOT NULL CHECK (status IN ('open', 'merged', 'closed')),
	UNIQUE (repo_id, repo_local_id),
	UNIQUE (repo_id, source_ref, destination_branch)
);

CREATE OR REPLACE FUNCTION create_repo_mr_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'repo_mr_seq_' || NEW.id;
BEGIN
	EXECUTE format('CREATE SEQUENCE %I', seq_name);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER after_insert_repo
AFTER INSERT ON repos
FOR EACH ROW
EXECUTE FUNCTION create_repo_mr_sequence();

CREATE OR REPLACE FUNCTION drop_repo_mr_sequence()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'repo_mr_seq_' || OLD.id;
BEGIN
	EXECUTE format('DROP SEQUENCE IF EXISTS %I', seq_name);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER before_delete_repo
BEFORE DELETE ON repos
FOR EACH ROW
EXECUTE FUNCTION drop_repo_mr_sequence();


CREATE OR REPLACE FUNCTION assign_repo_local_id()
RETURNS TRIGGER AS $$
DECLARE
	seq_name TEXT := 'repo_mr_seq_' || NEW.repo_id;
BEGIN
	IF NEW.repo_local_id IS NULL THEN
		EXECUTE format('SELECT nextval(%L)', seq_name)
		INTO NEW.repo_local_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER before_insert_merge_request
BEFORE INSERT ON merge_requests
FOR EACH ROW
EXECUTE FUNCTION assign_repo_local_id();
