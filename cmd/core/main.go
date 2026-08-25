package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luisfernandogaido/redistic/es"
	"github.com/luisfernandogaido/redistic/model"
	"github.com/redis/go-redis/v9"
)

type Meta struct {
	Index    string `json:"index"`
	Complete bool   `json:"complete"`
}

type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
	App       string    `json:"app"`
	Hostname  string    `json:"hostname"`
	Data      any       `json:"data"`
}

type MMD struct {
	Meta     Meta     `json:"meta"`
	Metadata Metadata `json:"data"`
}

type IndexData struct {
	Index string `json:"index"`
	Data  any    `json:"data"`
}

var (
	chIndexData = make(chan IndexData, 1000)
)

func init() {

}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go roteia()
	go despachaElastic()
	<-sigs

}

func roteia() {
	for {
		lote, err := extrai(1000)
		if err != nil {
			log.Println(err)
		}
		for _, message := range lote {
			s := message.(string)
			var mmd MMD
			if err := json.Unmarshal([]byte(s), &mmd); err != nil {
				log.Println(err)
				continue
			}
			if mmd.Meta.Complete {
				chIndexData <- IndexData{
					Index: mmd.Meta.Index,
					Data:  mmd.Metadata,
				}
			} else {
				fmt.Println("redis")
				fmt.Println(mmd)
			}
		}
	}
}

func despachaElastic() {
	indices := make([]string, 0)
	documentos := make([]any, 0)
	const lote = 1000
	for {
		select {
		case indexData := <-chIndexData:
			indices = append(indices, indexData.Index)
			documentos = append(documentos, indexData.Data)
			fmt.Println(indexData.Index, len(indices))
			if len(indices) == lote {
				if err := es.BulkIndexes(indices, documentos); err != nil {
					log.Println(err)
					time.Sleep(time.Second * 5)
				}
				indices = indices[:0]
				documentos = documentos[:0]
			}
		default:
			fmt.Println("esperando")
			time.Sleep(1 * time.Second)

		}
	}
}

func extrai(n int) ([]any, error) {
	lote := make([]any, 0, n)
	result, err := model.RedisCore.BLPop(model.Ctx, 0, "applog").Result()
	if err != nil {
		return nil, fmt.Errorf("extrai 1: %w", err)
	}
	lote = append(lote, result[1])
	messages, err := model.RedisCore.LPopCount(model.Ctx, "applog", n-1).Result()
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
