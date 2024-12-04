CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    song_name TEXT NOT NULL,
    group_name TEXT NOT NULL,
    release_date DATE NOT NULL,
    link TEXT NOT NULL
);


CREATE INDEX idx_group_name_song_name ON songs (group_name, song_name);
CREATE INDEX idx_song_name_group_name ON songs (song_name, group_name);
