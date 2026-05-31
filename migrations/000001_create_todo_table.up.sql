CREATE OR REPLACE FUNCTION notify_event()
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $$
    DECLARE 
        data jsonb;
        notification jsonb;

    BEGIN
        IF (TG_OP = 'DELETE') THEN
            data = to_jsonb(OLD);
        ELSE 
            data = to_jsonb(NEW);
        END IF;

        notification = jsonb_build_object(
            'table',
            TG_TABLE_NAME,
            'action',
            TG_OP,
            'timestamp',
            NOW(),
            'record',
            data
        );

        BEGIN
                PERFORM pg_notify('events', notification::text);
            EXCEPTION WHEN OTHERS THEN
                RAISE WARNING 'Notification failed: %', SQLERRM;
        END;

        RETURN NULL;
    END;
$$;

CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TRIGGER todos_notify_event
AFTER INSERT OR UPDATE OR DELETE ON todos
    FOR EACH ROW EXECUTE FUNCTION notify_event();