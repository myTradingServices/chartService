CREATE TYPE price_type AS ENUM ('ask', 'bid');
CREATE TABLE candles (
    symbol VARCHAR(30),
    bid_or_ask price_type,
    highest_price DECIMAL,
    lowest_price DECIMAL,
    open_price DECIMAL,
    close_pirce DECIMAL,
    open_time TIMESTAMP WITH TIME ZONE,
    time_interval INTERVAL,
    PRIMARY KEY(open_time)
);