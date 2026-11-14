ALTER TABLE todo_schema.todos
    RENAME COLUMN description TO content;

UPDATE todo_schema.todos
SET content = ''
WHERE content IS NULL;

ALTER TABLE todo_schema.todos
    ALTER COLUMN content SET NOT NULL;
