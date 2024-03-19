package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/chartService/internal/model"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	conn Interface

	input1 = model.Candle{
		Symbol:   "symb1",
		BidOrAsk: model.Bid,
		Highest:  decimal.New(15, 0),
		Lowest:   decimal.New(3, 0),
		Open:     decimal.New(4, 0),
		Close:    decimal.New(5, 0),
		OpenTime: time.Now(),
		Interval: time.Minute,
	}
	input2 = model.Candle{
		Symbol:   "symb1",
		BidOrAsk: model.Bid,
		Highest:  decimal.New(123, 0),
		Lowest:   decimal.New(33, 0),
		Open:     decimal.New(43, 0),
		Close:    decimal.New(53, 0),
		OpenTime: time.Now().Add(time.Second),
		Interval: time.Minute,
	}
)

func TestMain(m *testing.M) {
	ctx, _ := context.WithCancel(context.Background())

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:   "postgres_test",
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=user",
			"POSTGRES_DB=chart",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	postgresHostAndPort := pgResource.GetHostPort("5432/tcp")
	postgresUrl := fmt.Sprintf("postgres://user:password@%s/chart?sslmode=disable", postgresHostAndPort)

	log.Info("Connecting to database on url: ", postgresUrl)

	var dbpool *pgxpool.Pool
	if err = pool.Retry(func() error { // remove retry? (not nessesary)
		dbpool, err = pgxpool.New(ctx, postgresUrl)
		if err != nil {
			dbpool.Close()
			log.Error("can't connect to the pgxpool: %w", err)
		}
		return dbpool.Ping(ctx)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	commandArr := []string{
		"-url=jdbc:postgresql://" + postgresHostAndPort + "/chart",
		"-user=user",
		"-password=password",
		"-locations=filesystem:../../migrations/",
		"-schemas=trading",
		"-connectRetries=60",
		"migrate",
	}
	cmd := exec.Command("flyway", commandArr[:]...)

	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Printf("error: %s", err))
	}

	pool.MaxWait = 120 * time.Second
	conn = New(dbpool)

	code := m.Run()

	if err := pool.Purge(pgResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestAdd(t *testing.T) {
	testCtx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		input    model.Candle
		hasError bool
	}{
		{
			name:     "standart input-1",
			input:    input1,
			hasError: false,
		},
		{
			name:     "standart input-2",
			input:    input2,
			hasError: false,
		},
	}

	for _, test := range testTable {
		err := conn.Add(testCtx, test.input)
		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			assert.Nil(t, err, test.name)
		}
	}
	log.Info("TestAdd Finished!")
}

func TestGet(t *testing.T) {
	testCtx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		symbol   string
		interval time.Duration
		bora     model.PriceType
		actual   []model.Candle
		hasError bool
	}{
		{
			name:     "standart input with bid's & symb1",
			symbol:   input1.Symbol,
			interval: input1.Interval,
			bora:     input1.BidOrAsk,
			actual:   []model.Candle{input1, input2},
		},
	}

	for _, test := range testTable {
		expected, err := conn.Get(testCtx, test.symbol, test.interval, test.bora)
		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			if ok := assert.Nil(t, err, test.name); !ok {
				t.Error("Error is not nil:", err)
				continue
			}

			for i := range test.actual {
				assert.Equal(t, expected[i].BidOrAsk, test.actual[i].BidOrAsk)
				assert.Equal(t, expected[i].Close, test.actual[i].Close)
				assert.Equal(t, expected[i].Highest, test.actual[i].Highest)
				assert.Equal(t, expected[i].Lowest, test.actual[i].Lowest)
				assert.Equal(t, expected[i].Open, test.actual[i].Open)
				assert.Equal(t, expected[i].Symbol, test.actual[i].Symbol)
				assert.Equal(t, expected[i].Interval, test.actual[i].Interval)
				assert.Equal(t, expected[i].BidOrAsk, test.actual[i].BidOrAsk)

				if expected[i].OpenTime.Truncate(time.Second).Compare(test.actual[i].OpenTime.Truncate(time.Second)) != 0 {
					t.Fatal("expected OpenTime not equals actual OpenTime")
				}
			}
		}
	}
	log.Info("TestGet Finished!")
}

func TestDelete(t *testing.T) { // test not cheking value deleted or not (pgxpool do not return error if not found)
	testCtx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		symbol   string
		bora     model.PriceType
		hasError bool
	}{
		{
			name:     "delete valid candle",
			symbol:   "symbol1",
			bora:     model.Bid,
			hasError: false,
		},
		{
			name:     "delete non-existing candle",
			symbol:   "symbol2",
			bora:     model.Ask,
			hasError: false,
		},
	}

	for _, test := range testTable {
		err := conn.Delete(testCtx, test.symbol, test.bora)
		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			assert.Nil(t, err, test.name)
		}
	}
	log.Info("TestDelete Finished!")
}
