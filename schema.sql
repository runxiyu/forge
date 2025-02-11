CREATE TABLE groups (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT NOT NULL UNIQUE 
);

CREATE TABLE repos (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE RESTRICT, -- I mean, should be CASCADE but deleting Git repos on disk also needs to be considered
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

CREATE TABLE emails (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	list_id INTEGER NOT NULL REFERENCES mailing_lists(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	sender TEXT NOT NULL,
	date TIMESTAMP NOT NULL,
	content BYTEA NOT NULL
);

CREATE TABLE merge_requests (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	repo_id INTEGER NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
	creator INTEGER REFERENCES users(id) ON DELETE SET NULL,
	source_ref TEXT NOT NULL,
	destination_branch TEXT NOT NULL,
	status TEXT NOT NULL CHECK (status IN ('open', 'merged', 'closed')),
	created_at TIMESTAMP NOT NULL,
	mailing_list_id INT UNIQUE REFERENCES mailing_lists(id) ON DELETE CASCADE
);

CREATE TABLE users (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	password_algorithm TEXT NOT NULL CHECK (password_algorithm in ('argon2id')),
	password TEXT NOT NULL
);

CREATE TABLE ssh_public_keys (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	UNIQUE (user_id, content)
);
