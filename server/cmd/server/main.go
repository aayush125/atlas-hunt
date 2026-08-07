package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/engine"
	"atlas.sageyutta.com/internal/validator"
	"atlas.sageyutta.com/internal/ws"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port          int
	env           string
	redisAddr     string
	redisPassword string
}

type application struct {
	config       config
	logger       *slog.Logger
	models       *data.Models
	rdb          *redis.Client
	wg           sync.WaitGroup
	mu           sync.Mutex
	shuttingDown bool
	gameRooms    map[string]*ws.Room
	engine       *engine.Engine
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.redisAddr, "redis-addr", "localhost:6379", "Redis address")
	flag.StringVar(&cfg.redisPassword, "redis-password", "", "Redis password")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if !validator.PermittedValue(cfg.env, "development", "staging", "production") {
		logger.Error("invalid app environment", "env", cfg.env)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.redisAddr,
		Password:     cfg.redisPassword,
		DB:           0,
		Protocol:     2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("could not connect to redis", "error", err)
		os.Exit(1)
	}

	app := &application{
		config:       cfg,
		models:       data.NewModels(rdb),
		rdb:          rdb,
		logger:       logger,
		shuttingDown: false,
		gameRooms:    make(map[string]*ws.Room),
		engine:       engine.NewEngine(logger),
	}

	err := app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
