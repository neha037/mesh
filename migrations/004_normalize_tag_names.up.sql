-- Normalize existing tag names to lowercase
UPDATE tags SET name = lower(trim(name));
