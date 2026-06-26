-- student_id is DB-assigned (GENERATED ALWAYS AS IDENTITY); do not specify it.
INSERT INTO students (name, address, grade) VALUES
    ('Ada Lovelace',      '12 Analytical Way, London',    10),
    ('Alan Turing',       '7 Enigma Road, Bletchley',     11),
    ('Grace Hopper',      '34 Compiler Court, New York',  12),
    ('Katherine Johnson', '88 Orbit Lane, Hampton',       10),
    ('Dennis Ritchie',    '5 Unix Street, Murray Hill',   11);
