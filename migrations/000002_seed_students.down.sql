-- IDs are DB-assigned, so revert by the known seed names.
DELETE FROM students WHERE name IN (
    'Ada Lovelace',
    'Alan Turing',
    'Grace Hopper',
    'Katherine Johnson',
    'Dennis Ritchie'
);
