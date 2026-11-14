ALTER TABLE todo_schema.todos
    ALTER COLUMN content DROP NOT NULL;

ALTER TABLE todo_schema.todos
    RENAME COLUMN content TO description;
