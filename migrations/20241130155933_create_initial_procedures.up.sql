CREATE OR REPLACE FUNCTION add_song(
    song_name_ TEXT,
    group_name_ TEXT,
    release_date_ DATE,
    link_ TEXT,
    verses TEXT[]
) RETURNS INT AS $$
DECLARE
    new_song_id INT;
BEGIN
    INSERT INTO songs (song_name, group_name, release_date, link)
    VALUES ($1, $2, $3, $4)
        RETURNING id INTO new_song_id;

    FOR i IN 1..array_length(verses, 1) LOOP
        INSERT INTO song_lyrics (song_id, verse_number, verse_text)
        VALUES (new_song_id, i, verses[i]);
    END LOOP;

    RETURN new_song_id;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE PROCEDURE update_song_info(
    id_ INT,
    song_name_ TEXT,
    group_name_ TEXT,
    release_date_ DATE,
    link_ TEXT
) AS $$
BEGIN
    UPDATE songs
    SET song_name = CASE WHEN $2 <> '' THEN $2 ELSE song_name END,
        group_name = CASE WHEN $3 <> '' THEN $3 ELSE group_name END,
        release_date = CASE WHEN $4 <> '0001-01-01' THEN $4 ELSE release_date END,
        link = CASE WHEN $5 <> '' THEN $5 ELSE link END
    WHERE id = $1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Song with id % not found', id_;
    END IF;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_song_verses(
    id_ INT,
    offset_verse INT,
    limit_verse INT
) RETURNS TABLE(verse_text_ TEXT) AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM songs
        WHERE id = $1
    ) THEN
        RAISE EXCEPTION 'Song with id % not found', id_;
        RETURN;
    END IF;

    RETURN QUERY
    SELECT verse_text
    FROM song_lyrics
    WHERE song_id = $1
      AND verse_number > $2
    ORDER BY verse_number
    LIMIT $3;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_songs_info(
    prev_song TEXT,
    prev_group TEXT,
    limit_verse INT
) RETURNS SETOF songs AS $$
BEGIN
    RETURN QUERY
    SELECT *
    FROM songs
    WHERE song_name >= $1
      AND group_name > $2
    ORDER BY song_name, group_name
    LIMIT $3;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_group_songs_info(
    group_name_ TEXT,
    prev_song TEXT,
    limit_verse INT
) RETURNS SETOF songs AS $$
BEGIN
    RETURN QUERY
    SELECT *
    FROM songs
    WHERE group_name = $1
      AND song_name > $2
    ORDER BY song_name
    LIMIT $3;
END;
$$ LANGUAGE plpgsql;