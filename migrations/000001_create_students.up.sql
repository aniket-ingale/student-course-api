CREATE TABLE students (
    student_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT    NOT NULL,
    address    TEXT    NOT NULL,
    grade      INTEGER NOT NULL
);
