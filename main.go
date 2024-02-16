package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/chartService/internal/config"
	"github.com/mmfshirokan/chartService/internal/repository"
	"github.com/mmfshirokan/chartService/internal/rpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	conf, err := config.New()
	if err != nil {
		log.Errorf("Error occurred while parsing config: %v", err)
		return
	}

	option := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(conf.RpcURI, option)
	if err != nil {
		log.Errorf("grpc connection error on %v: %v", conf.RpcURI, err)
		return
	}
	defer conn.Close()

	dbpool, err := pgxpool.New(ctx, conf.PostgresURI)
	if err != nil {
		log.Errorf("Error occurred while connecting yo postgresql pool: %v", err)
		return
	}

	repo := repository.New(dbpool)
	dataStream := rpc.New(conn, repo)

	forever := make(chan struct{})
	go dataStream.Receive()
	<-forever
}
