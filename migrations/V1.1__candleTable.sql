CREATE TABLE trading.candles (
    symbol VARCHAR(30),
    bid_or_ask INT,
    highest_price DECIMAL,
    lowest_price DECIMAL,
    open_price DECIMAL,
    close_pirce DECIMAL,
    open_time TIMESTAMP WITH TIME ZONE,
    time_interval INTERVAL,
    PRIMARY KEY(open_time)
);