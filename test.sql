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
)
INSERT INTO repos (name, group_id, contrib_requirements, filesystem_path)
SELECT 'server', id, 'public', '/home/runxiyu/Lindenii/forge/server/.git'
FROM child_group;

SELECT * FROM groups;
SELECT * FROM repos;
