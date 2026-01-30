-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 
     name TEXT NOT NULL,
     start_time TIMESTAMP NOT NULL,
 
     total_seats INT NOT NULL,
     available_seats INT NOT NULL,
 
     created_at TIMESTAMP DEFAULT NOW()
 );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table events;
-- +goose StatementEnd
