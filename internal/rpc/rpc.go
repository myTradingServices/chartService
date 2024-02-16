package rpc

import (
	"context"
	"io"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/chartService/internal/model"
	"github.com/mmfshirokan/chartService/internal/repository"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type server struct {
	conn   *grpc.ClientConn
	candle repository.Cnadle
}

type Receiver interface {
	Receive()
}

func New(connection *grpc.ClientConn, candle repository.Cnadle) Receiver {
	return &server{
		conn:   connection,
		candle: candle,
	}
}

func (serv *server) Receive() {
	ctx := context.Background()
	consumer := pb.NewConsumerClient(serv.conn)

	stream, err := consumer.DataStream(ctx, &pb.RequestDataStream{Start: true})
	if err != nil {
		log.Errorf("Error in DataStream: %v", err)
		return
	}
	defer stream.CloseSend()

	BidCandle := newEmptyCandle("", "Bid")
	AskCandle := newEmptyCandle("", "Ask")

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Infof("Exitin stream, because error is %v", err)
			break
		}
		if err != nil {
			log.Errorf("Error occured: %v", err)
			return
		}

		if BidCandle.Symbol == "" {
			BidCandle.Symbol = recv.Symbol
		}
		if AskCandle.Symbol == "" {
			AskCandle.Symbol = recv.Symbol
		}

		if BidCandle.OpenTime.IsZero() {
			BidCandle.OpenTime = recv.Date.AsTime()
			BidCandle.Open = decimal.New(recv.Bid.Value, recv.Bid.Exp)
		}
		if AskCandle.OpenTime.IsZero() {
			AskCandle.OpenTime = recv.Date.AsTime()
			AskCandle.Open = decimal.New(recv.Ask.Value, recv.Ask.Exp)
		}

		if price := decimal.New(recv.Bid.Value, recv.Bid.Exp); BidCandle.Highest.LessThan(price) || BidCandle.Highest.IsZero() {
			BidCandle.Highest = price
		}
		if price := decimal.New(recv.Ask.Value, recv.Ask.Exp); AskCandle.Highest.LessThan(price) || AskCandle.Highest.IsZero() {
			AskCandle.Highest = price
		}

		if price := decimal.New(recv.Bid.Value, recv.Bid.Exp); BidCandle.Lowest.GreaterThan(price) || BidCandle.Lowest.IsZero() {
			BidCandle.Lowest = price
		}
		if price := decimal.New(recv.Ask.Value, recv.Ask.Exp); AskCandle.Lowest.GreaterThan(price) || AskCandle.Lowest.IsZero() {
			AskCandle.Lowest = price
		}

		if time.Since(BidCandle.OpenTime) >= time.Minute {
			BidCandle.CloseTime = time.Now()
			BidCandle.Close = decimal.New(recv.Bid.Value, recv.Bid.Exp)

			err = serv.candle.Add(ctx, BidCandle)
			if err != nil {
				log.Errorf("SQL error occured: %v", err)
				return
			}

			BidCandle = newEmptyCandle(recv.Symbol, "Bid")
		}
		if time.Since(AskCandle.OpenTime) >= time.Minute {
			AskCandle.CloseTime = time.Now()
			AskCandle.Close = decimal.New(recv.Ask.Value, recv.Ask.Exp)

			err = serv.candle.Add(ctx, AskCandle)
			if err != nil {
				log.Errorf("SQL error occured: %v", err)
				return
			}

			AskCandle = newEmptyCandle(recv.Symbol, "Ask")
		}
	}
}

func newEmptyCandle(symbol, bidOrAsk string) model.Candle {
	return model.Candle{
		Symbol:    symbol,
		BidOrAsk:  bidOrAsk,
		Highest:   decimal.Zero,
		Lowest:    decimal.Zero,
		Open:      decimal.Zero,
		Close:     decimal.Zero,
		OpenTime:  time.Time{},
		CloseTime: time.Time{},
	}
}
