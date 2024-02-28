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

func (reciver *server) Receive(ctx context.Context, interval time.Duration) {
	consumer := pb.NewConsumerClient(reciver.conn)

	stream, err := consumer.DataStream(ctx, &pb.RequestDataStream{Start: true})
	if err != nil {
		log.Errorf("Error in DataStream: %v", err)
		return
	}
	defer stream.CloseSend()

	symbolMap := make(map[string]chan model.Price)

	//counter := 1 // delete
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

		ch, ok := symbolMap[recv.Symbol]
		if !ok {
			ch = make(chan model.Price)
			symbolMap[recv.Symbol] = ch

			go streamHandler(ctx, ch, interval, reciver.serv)

			ch <- model.Price{
				Symbol: recv.Symbol,
				Bid:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
				Ask:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
				Date:   recv.Date,
			}

			continue
		}

		ch <- model.Price{
			Symbol: recv.Symbol,
			Bid:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
			Ask:    model.Decim{Value: recv.Bid.Value, Exp: recv.Bid.Exp},
			Date:   recv.Date,
		}
		// log.Info("Data received: ", counter) // delete
		// counter++                            // delete
	}
}

func streamHandler(ctx context.Context, ch chan model.Price, interval time.Duration, serv service.Interface) {
	recv := <-ch
	BidCandle := model.Candle{
		Symbol:   recv.Symbol,
		BidOrAsk: model.Bid,
		Open:     decimal.New(recv.Bid.Value, recv.Ask.Exp),
		OpenTime: recv.Date.AsTime(),
		Interval: interval,
	}
	AskCandle := model.Candle{
		Symbol:   recv.Symbol,
		BidOrAsk: model.Ask,
		Open:     decimal.New(recv.Bid.Value, recv.Ask.Exp),
		OpenTime: recv.Date.AsTime(),
		Interval: interval,
	}

	candlesAreEmpty := false

	for {
		if candlesAreEmpty {
			BidCandle = model.Candle{
				Open:     decimal.New(recv.Bid.Value, recv.Ask.Exp),
				OpenTime: recv.Date.AsTime(),
				Highest:  decimal.Zero,
				Lowest:   decimal.Zero,
			}
			AskCandle = model.Candle{
				Open:     decimal.New(recv.Bid.Value, recv.Ask.Exp),
				OpenTime: recv.Date.AsTime(),
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
				log.Errorf("SQL error occured: %v", err)
				return
			}

			AskCandle.Close = decimal.New(recv.Ask.Value, recv.Ask.Exp)
			// //tmp
			// log.Info("bidorask: ", AskCandle.BidOrAsk, " lowest: ", AskCandle.Lowest.String(), " highest: ", AskCandle.Highest.String(), " close: ", AskCandle.Close.String(), " interval: ", AskCandle.Interval, " symbol: ", AskCandle.Symbol, " open: ", AskCandle.Open.String())
			// //tmp
			err = serv.Add(ctx, AskCandle)
			if err != nil {
				log.Errorf("SQL error occured: %v", err)
				return
			}

			candlesAreEmpty = true
		}
		recv = <-ch
	}

}
