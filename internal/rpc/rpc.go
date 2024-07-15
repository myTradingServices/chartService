package rpc

import (
	"context"
	"io"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/chartService/internal/model"
	"github.com/mmfshirokan/chartService/internal/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// TODO change model decim to shopspring/decimal

type server struct {
	conn *grpc.ClientConn
	serv service.Interface
}

type Receiver interface {
	Receive(context.Context, time.Duration)
}

func New(connection *grpc.ClientConn, serv service.Interface) Receiver {
	return &server{
		conn: connection,
		serv: serv,
	}
}

func (r *server) Receive(ctx context.Context, interval time.Duration) {
	log.Info("Rpc recive started")

	consumer := pb.NewConsumerClient(r.conn)

	stream, err := consumer.DataStream(ctx, &pb.RequestDataStream{Start: true})
	if err != nil {
		log.Errorf("Error in DataStream attempting to reconnect in 15 seconds: %v", err)
		time.Sleep(time.Second * 15)

		stream, err = consumer.DataStream(ctx, &pb.RequestDataStream{Start: true})
		if err != nil {
			log.Errorf("Reconnection failed exiting: %v", err)
			return
		}
	}
	defer stream.CloseSend()

	symbolMap := make(map[string]chan model.Price)

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Infof("Exitin stream, because error is %v", err)
			break
		}
		if err != nil {
			log.Errorf("Error occured: %v, exiting sream", err)
			return
		}

		ch, ok := symbolMap[recv.Symbol]
		if !ok {
			log.Info("New symbol handling started: ", recv.Symbol)

			ch = make(chan model.Price)
			symbolMap[recv.Symbol] = ch

			go streamHandler(ctx, ch, interval, r.serv)

			log.Info("Sending price for the first time trough ch for ", recv.Symbol)
			ch <- model.Price{
				Symbol: recv.Symbol,
				Bid:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
				Ask:    model.Decim{Value: recv.Ask.Value, Exp: recv.Ask.Exp},
				Date:   recv.Date,
			}
			log.Infof("Sending price for the first time trough ch for %v completed", recv.Symbol)

			continue
		}

		log.Info("Sending trough ch price for ", recv.Symbol)
		ch <- model.Price{
			Symbol: recv.Symbol,
			Bid:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
			Ask:    model.Decim{Value: recv.Ask.Value, Exp: recv.Ask.Exp},
			Date:   recv.Date,
		}
		log.Infof("Sending price trough ch for %v completed", recv.Symbol)

	}
}

func streamHandler(ctx context.Context, ch chan model.Price, interval time.Duration, serv service.Interface) {
	log.Info("Stream handler started; first recv price from ch")
	recv := <-ch
	log.Info("First recv price from ch completed (stream handler)")

	BidCandle := model.Candle{
		Symbol:   recv.Symbol,
		BidOrAsk: model.Bid,
		Open:     decimal.New(recv.Bid.Value, recv.Bid.Exp),
		OpenTime: recv.Date.AsTime(),
		Interval: interval,
	}
	AskCandle := model.Candle{
		Symbol:   recv.Symbol,
		BidOrAsk: model.Ask,
		Open:     decimal.New(recv.Ask.Value, recv.Ask.Exp),
		OpenTime: recv.Date.AsTime(),
		Interval: interval,
	}

	candlesAreEmpty := false

	for {
		if candlesAreEmpty {
			BidCandle = model.Candle{
				Symbol:   recv.Symbol,
				Open:     decimal.New(recv.Bid.Value, recv.Bid.Exp),
				BidOrAsk: model.Bid,
				OpenTime: recv.Date.AsTime(),
				Interval: interval,
				Highest:  decimal.Zero,
				Lowest:   decimal.Zero,
			}
			AskCandle = model.Candle{
				Symbol:   recv.Symbol,
				Open:     decimal.New(recv.Ask.Value, recv.Ask.Exp),
				BidOrAsk: model.Ask,
				OpenTime: recv.Date.AsTime(),
				Interval: interval,
				Highest:  decimal.Zero,
				Lowest:   decimal.Zero,
			}

			candlesAreEmpty = false
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

		if time.Since(BidCandle.OpenTime) >= interval && time.Since(AskCandle.OpenTime) >= interval {
			BidCandle.Close = decimal.New(recv.Bid.Value, recv.Bid.Exp)
			err := serv.Add(ctx, BidCandle)
			if err != nil {
				log.Errorf("SQL error occured while attempting to save bid candle: %v", err)
				//break //return
			} else {
				log.Info("Bid candle Added to DB")
			}

			AskCandle.Close = decimal.New(recv.Ask.Value, recv.Ask.Exp)
			err = serv.Add(ctx, AskCandle)
			if err != nil {
				log.Errorf("SQL error occured while attempting to save ask candle: %v", err)
				//break //return
			} else {
				log.Info("Ask candle Added to DB")
			}

			candlesAreEmpty = true
		}

		log.Info("Stream handler recv new price...")
		recv = <-ch
		log.Info("Stream handler recv new price completed")
	}

}
