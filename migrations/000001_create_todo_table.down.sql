DROP FUNCTION IF EXISTS notify_event();
DROP TABLE IF EXISTS todos;
DROP TRIGGER IF EXISTS todos_notify_event ON todos;