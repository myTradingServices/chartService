CREATE SCHEMA trading CREATE TABLE candles (
    symbol VARCHAR(30),
    bid_or_ask CHAR(3),
    highest_price DECIMAL,
    lowest_price DECIMAL,
    open_price DECIMAL,
    close_pirce DECIMAL,
    open_time TIMESTAMP,
    close_time TIMESTAMP,
    PRIMARY KEY(open_time, close_time)
)