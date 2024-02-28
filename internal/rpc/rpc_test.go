package rpc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	"github.com/mmfshirokan/chartService/internal/model"
	mocks "github.com/mmfshirokan/chartService/internal/rpc/mock"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	target string = "localhost:2021"
)

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", target)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
	}

	rpcServer := grpc.NewServer()
	testSonsumer := NewTestConsumerServer()
	pb.RegisterConsumerServer(rpcServer, testSonsumer)

	go func() {
		err = rpcServer.Serve(lis)
		if err != nil {
			log.Error("rpc fatal error: Server can't start")
			return
		}
	}()

	code := m.Run()

	os.Exit(code)
}

func TestReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*70)
	defer cancel()

	option := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(target, option)
	if err != nil {
		log.Errorf("grpc connection error on %v: %v", target, err)
		return
	}
	defer conn.Close()

	mocker := mocks.NewInterface(t)
	recv := New(conn, mocker)

	mocker.EXPECT().Add(ctx, mock.MatchedBy(func(c model.Candle) bool {
		return c.Symbol == "symb1" &&
			c.BidOrAsk == model.Bid &&
			c.Open.Equal(decimal.New(0, 0)) &&
			c.Close.Equal(decimal.New(60, 0)) &&
			c.Interval == time.Minute &&
			c.Highest.Equal(decimal.New(60, 0)) &&
			c.Lowest.Equal(decimal.New(1, 0))
	})).Return(nil)

	mocker.EXPECT().Add(ctx, mock.MatchedBy(func(c model.Candle) bool {
		return c.BidOrAsk == model.Ask &&
			c.Open.Equal(decimal.New(1, 0)) &&
			c.Close.Equal(decimal.New(61, 0)) &&
			c.Interval == time.Minute &&
			c.Highest.Equal(decimal.New(61, 0)) &&
			c.Lowest.Equal(decimal.New(1, 0))
	})).Return(nil)

	go recv.Receive(ctx, time.Minute)

	time.Sleep(time.Second * 65)

	mocker.AssertExpectations(t)
}
