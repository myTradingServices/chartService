package rpc

import (
	"errors"
	"io"
	"time"

	"github.com/mmfshirokan/PriceService/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testServer struct {
	pb.UnimplementedConsumerServer
}

func NewTestConsumerServer() pb.ConsumerServer {
	return &testServer{}
}

func (t *testServer) DataStream(req *pb.RequestDataStream, stream pb.Consumer_DataStreamServer) error {
	if !req.Start {
		return errors.New("start is not initiated")
	}
	for i := 0; i < 62; i++ {
		err := stream.Send(&pb.ResponseDataStream{
			Date: timestamppb.New(time.Now()),
			Bid: &pb.ResponseDataStreamDecimal{
				Value: int64(i),
				Exp:   0,
			},
			Ask: &pb.ResponseDataStreamDecimal{
				Value: int64(i + 1),
				Exp:   0,
			},
			Symbol: "symb1",
		})
		if err == io.EOF {
			log.Infof("Stream exited, because error is: %v", err)
			break
		}
		if err != nil {
			log.Errorf("Error sending message: %v.", err)
		}

		// log.Infof("Sent %v", i) // delete
		time.Sleep(time.Second)
	}

	return nil
}
