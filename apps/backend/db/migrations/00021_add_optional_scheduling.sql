-- +goose Up
ALTER TABLE tournaments ADD COLUMN venue text NOT NULL DEFAULT '' CHECK (char_length(venue) <= 160);
ALTER TABLE matches
 ADD COLUMN scheduled_date text NOT NULL DEFAULT '',
 ADD COLUMN scheduled_time text NOT NULL DEFAULT '',
 ADD COLUMN time_zone text NOT NULL DEFAULT '',
 ADD COLUMN venue text NOT NULL DEFAULT '' CHECK (char_length(venue) <= 160),
 ADD CONSTRAINT match_schedule_date CHECK (scheduled_date = '' OR scheduled_date ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'),
 ADD CONSTRAINT match_schedule_time CHECK (scheduled_time = '' OR scheduled_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
 ADD CONSTRAINT match_schedule_zone CHECK ((scheduled_time = '' AND time_zone = '') OR (scheduled_time <> '' AND char_length(time_zone) BETWEEN 1 AND 100));

-- +goose Down
ALTER TABLE matches DROP CONSTRAINT match_schedule_zone, DROP CONSTRAINT match_schedule_time, DROP CONSTRAINT match_schedule_date,
 DROP COLUMN venue, DROP COLUMN time_zone, DROP COLUMN scheduled_time, DROP COLUMN scheduled_date;
ALTER TABLE tournaments DROP COLUMN venue;
