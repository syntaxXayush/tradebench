package main

import (
"context"
"log"
"log/slog"
"net/http"
"os"

"github.com/bench/api-gateway/config"
"github.com/bench/api-gateway/handlers"
"github.com/bench/api-gateway/store"
)

func main() {
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

cfg := config.Load()
ctx := context.Background()

postgresStore, err := store.NewPostgresStore(ctx, cfg.PostgresDSN)
if err != nil {
log.Fatalf("FATAL: postgres init failed: %v", err)
}

redisClient, err := store.NewRedisClient(cfg.RedisAddr)
if err != nil {
log.Fatalf("FATAL: redis init failed: %v", err)
}
defer redisClient.Close()

mux := http.NewServeMux()
handlers.NewSubmissionHandler(cfg, postgresStore, redisClient).Register(mux)
handlers.NewLeaderboardHandler(cfg, postgresStore, redisClient).Register(mux)
handlers.NewAdminHandler(cfg, postgresStore, redisClient).Register(mux)

slog.Info("api-gateway starting", "addr", ":8080")
if err := http.ListenAndServe(":8080", mux); err != nil {
log.Fatal(err)
}
}
