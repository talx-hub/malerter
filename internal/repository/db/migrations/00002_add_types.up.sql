BEGIN TRANSACTION;

INSERT INTO type(name_type)
    VALUES
    ('gauge'),
    ('counter');

COMMIT;
