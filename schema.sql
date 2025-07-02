CREATE TABLE IF NOT EXISTS cached_puzzles (
    id SERIAL PRIMARY KEY,
    request_hash TEXT UNIQUE NOT NULL,
    request_params JSONB NOT NULL,
    response_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
