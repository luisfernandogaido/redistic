package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/luisfernandogaido/redistic/model"
	"github.com/redis/go-redis/v9"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go transfere()
	<-sigs
}

func transfere() {
	for {
		lote, err := extrai(10000)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := envia(lote); err != nil {
			fmt.Println(err)
			slices.Reverse(lote)
			if err := model.RedisEdge.LPush(model.Ctx, "applog", lote...).Err(); err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 5)
			continue
		}
		fmt.Println(len(lote))
	}
}

func extrai(n int) ([]any, error) {
	lote := make([]any, 0, n)
	result, err := model.RedisEdge.BLPop(model.Ctx, 0, "applog").Result()
	if err != nil {
		return nil, fmt.Errorf("extrai 1: %w", err)
	}
	lote = append(lote, result[1])
	messages, err := model.RedisEdge.LPopCount(model.Ctx, "applog", n-1).Result()
	if errors.Is(err, redis.Nil) {
		return lote, nil
	}
	if err != nil {
		return nil, fmt.Errorf("extrai 2: %w", err)
	}
	for _, message := range messages {
		lote = append(lote, message)
	}
	return lote, nil
}

func envia(lote []any) error {
	err := model.RedisCore.RPush(model.Ctx, "applog", lote...).Err()
	if err != nil {
		return fmt.Errorf("envia: %w", err)
	}
	return nil
}
