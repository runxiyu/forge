WITH parent_group AS (
	INSERT INTO groups (name, description)
	VALUES ('lindenii', 'The Lindenii Project')
	RETURNING id
),
child_group AS (
	INSERT INTO groups (name, description, parent_group)
	SELECT 'forge', 'Lindenii Forge', id
	FROM parent_group
	RETURNING id
),
create_repos AS (
	INSERT INTO repos (name, group_id, contrib_requirements, filesystem_path)
	SELECT 'server', id, 'public', '/home/runxiyu/Lindenii/forge/server/.git'
	FROM child_group
),
new_user AS (
	INSERT INTO users (username, type, password)
	VALUES ('test', 'registered', '$argon2id$v=19$m=4096,t=3,p=1$YWFhYWFhYWFhYWFh$i40k7TPFHqXRH4eQOAYGH3LvzwQ38jqqlfap9Rtiy3c')
	RETURNING id
),
new_ssh AS (
	INSERT INTO ssh_public_keys (key_string, user_id)
	SELECT 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAuavKDhEM1L6CufIecy2P712gp151CqZuwSYahTWvmq', id
	FROM new_user
	RETURNING user_id
)
INSERT INTO user_group_roles (group_id, user_id)
SELECT child_group.id, new_ssh.user_id
FROM child_group, new_ssh;

SELECT * FROM groups;
SELECT * FROM repos;
SELECT * FROM users;
SELECT * FROM ssh_public_keys;
SELECT * FROM user_group_roles;

