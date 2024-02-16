package rpc

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Errorf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Errorf("Could not connect to Docker: %s", err)
		return
	}

	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:   "postgres_test",
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=pgpw4echo",
			"POSTGRES_USER=echopguser",
			"POSTGRES_DB=echodb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Errorf("Could not start resource: %s", err)
		return
	}

	postgresHostAndPort := pgResource.GetHostPort("5432/tcp")
	postgresUrl := fmt.Sprintf("postgres://echopguser:pgpw4echo@%s/echodb?sslmode=disable", postgresHostAndPort)

	log.Info("Connecting to database on url: ", postgresUrl)

	var dbpool *pgxpool.Pool
	if err = pool.Retry(func() error {
		dbpool, err = pgxpool.New(ctx, postgresUrl)
		if err != nil {
			dbpool.Close()
			log.Errorf("can't connect to the pgxpool: %v", err)
		}
		return dbpool.Ping(ctx)
	}); err != nil {
		log.Errorf("Could not connect to docker: %s", err)
		return
	}

	commandArr := []string{
		"-url=jdbc:postgresql://" + postgresHostAndPort + "/echodb",
		"-user=echopguser",
		"-password=pgpw4echo",
		"-locations=filesystem:../../migrations/sql",
		"-schemas=apps", //remove?
		"-connectRetries=60",
		"migrate",
	}
	cmd := exec.Command("flyway", commandArr[:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Printf("error: %s", err))
	}
	log.Info(fmt.Printf("out: %s%s", outb.String(), errb.String()))

	// TODO: finish
}
