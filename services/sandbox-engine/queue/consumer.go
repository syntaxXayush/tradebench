package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/bench/sandbox-engine/runner"
)

const (
	jobStreamKey  = "stream:jobs"
	consumerGroup = "sandbox-engine"
	consumerName  = "sandbox-engine-1"
	blockDuration = 5 * time.Second
)

type StatusUpdater interface {
	UpdateStatus(ctx context.Context, submissionID, status, errMsg string) error
}

type Consumer struct {
	rdb     *redis.Client
	builder *runner.Builder
	spawner *runner.Spawner
	health  *runner.HealthChecker
	db      StatusUpdater
}

func NewConsumer(rdb *redis.Client, builder *runner.Builder, spawner *runner.Spawner, health *runner.HealthChecker, db StatusUpdater) *Consumer {
	return &Consumer{
		rdb:     rdb,
		builder: builder,
		spawner: spawner,
		health:  health,
		db:      db,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	// Ensure consumer group exists (ignore BUSYGROUP error if already created).
	err := c.rdb.XGroupCreateMkStream(ctx, jobStreamKey, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("consumer: create group: %w", err)
	}

	slog.Info("consumer: listening", "stream", jobStreamKey, "group", consumerGroup)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{jobStreamKey, ">"},
			Count:    1,
			Block:    blockDuration,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("consumer: xreadgroup error", "err", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if err := c.process(ctx, msg); err != nil {
					slog.Error("consumer: process failed", "msgId", msg.ID, "err", err)
					// Leave the message in PEL for redelivery; don't ACK.
					continue
				}
				// ACK on success.
				_ = c.rdb.XAck(ctx, jobStreamKey, consumerGroup, msg.ID).Err()
			}
		}
	}
}

func (c *Consumer) process(ctx context.Context, msg redis.XMessage) error {
	vals := msg.Values
	submissionID, _ := vals["submissionId"].(string)
	zipPath, _ := vals["zipPath"].(string)

	if submissionID == "" || zipPath == "" {
		return fmt.Errorf("consumer: missing submissionId or zipPath in message %s", msg.ID)
	}

	imageTag := "bench-submission-" + submissionID

	// BUILDING
	if err := c.db.UpdateStatus(ctx, submissionID, "BUILDING", ""); err != nil {
		slog.Warn("consumer: failed to set BUILDING status", "submissionId", submissionID, "err", err)
	}
	if err := c.builder.Build(zipPath, imageTag); err != nil {
		_ = c.db.UpdateStatus(ctx, submissionID, "FAILED", err.Error())
		return fmt.Errorf("consumer: build failed: %w", err)
	}

	// RUNNING
	if err := c.db.UpdateStatus(ctx, submissionID, "RUNNING", ""); err != nil {
		slog.Warn("consumer: failed to set RUNNING status", "submissionId", submissionID, "err", err)
	}
	containerID, _, err := c.spawner.Spawn(imageTag, submissionID)
	if err != nil {
		_ = c.db.UpdateStatus(ctx, submissionID, "FAILED", err.Error())
		return fmt.Errorf("consumer: spawn failed: %w", err)
	}

	if err := c.health.WaitReady(ctx, containerID); err != nil {
		_ = c.db.UpdateStatus(ctx, submissionID, "FAILED", "container failed health check: "+err.Error())
		return fmt.Errorf("consumer: health check failed: %w", err)
	}

	if err := c.db.UpdateStatus(ctx, submissionID, "BENCHMARKING", ""); err != nil {
		slog.Warn("consumer: failed to set BENCHMARKING status", "submissionId", submissionID, "err", err)
	}

	slog.Info("consumer: submission ready for benchmarking",
		"submissionId", submissionID,
		"containerID", containerID,
		"imageTag", imageTag,
		)
	return nil
}
