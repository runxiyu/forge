CREATE TABLE groups (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT
);

CREATE TABLE repos (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT, -- I mean, should be CASCADE but deleting Git repos on disk also needs to be considered
	contrib_requirements TEXT NOT NULL CHECK (contrib_requirements IN ('closed', 'registered_user', 'ssh_pubkey', 'public')),
	name TEXT NOT NULL,
	UNIQUE(group_id, name),
	description TEXT,
	filesystem_path TEXT
);

CREATE TABLE ticket_trackers (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	UNIQUE(group_id, name),
	description TEXT
);

CREATE TABLE tickets (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tracker_id INTEGER NOT NULL REFERENCES ticket_trackers(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT
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

CREATE TABLE users (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username TEXT UNIQUE,
	type TEXT NOT NULL CHECK (type IN ('pubkey_only', 'registered')),
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

-- TODO:
CREATE TABLE merge_requests (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	title TEXT,
	repo_id INTEGER NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
	creator INTEGER REFERENCES users(id) ON DELETE SET NULL,
	source_ref TEXT NOT NULL,
	destination_branch TEXT,
	status TEXT NOT NULL CHECK (status IN ('open', 'merged', 'closed')),
	UNIQUE (repo_id, source_ref, destination_branch),
	UNIQUE (repo_id, id)
);

CREATE TABLE user_group_roles (
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	PRIMARY KEY(user_id, group_id)
);
