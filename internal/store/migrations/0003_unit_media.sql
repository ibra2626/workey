-- Optional reference fields per unit: the source's own title for the unit and a
-- link to open it (a video URL). Both stay NULL for units that have neither.
ALTER TABLE units ADD COLUMN title TEXT;
ALTER TABLE units ADD COLUMN url TEXT;
