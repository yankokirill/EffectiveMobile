CREATE TABLE song_lyrics (
    song_id INT NOT NULL,
    verse_number INT NOT NULL,
    verse_text TEXT NOT NULL,
    CONSTRAINT fk_song FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
);

CREATE INDEX idx_song_lyrics_song_id_verse_number ON song_lyrics (song_id, verse_number);
