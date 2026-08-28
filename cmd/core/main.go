package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/luisfernandogaido/redistic/es"
	"github.com/luisfernandogaido/redistic/gd"
	"github.com/luisfernandogaido/redistic/model"
	"github.com/redis/go-redis/v9"
)

const (
	tamLote = 1000
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
	chElastic = make(chan IndexData, tamLote)
	chRedis   = make(chan IndexData, tamLote)
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go roteia()
	go despachaElastic()
	go despachaRedis()
	go preprocessaNginx()
	<-sigs
}

func roteia() {
	for {
		lote, err := extrai(tamLote)
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
				chElastic <- IndexData{
					Index: mmd.Meta.Index,
					Data:  mmd.Metadata,
				}
			} else {
				chRedis <- IndexData{
					Index: mmd.Meta.Index,
					Data:  mmd.Metadata,
				}
			}
		}
	}
}

func despachaElastic() {
	tempoMaximoEspera := 5 * time.Second
	indices := make([]string, 0, tamLote)
	documentos := make([]any, 0, tamLote)
	ticker := time.NewTicker(tempoMaximoEspera)
	defer ticker.Stop()

	flush := func() {
		if len(indices) == 0 {
			return
		}
		if err := es.BulkIndexes(indices, documentos); err != nil {
			log.Printf("erro ao enviar lote: %v\n", err)
			time.Sleep(tempoMaximoEspera) // backoff simples
		}
		indices = indices[:0]
		documentos = documentos[:0]
	}

	for {
		select {
		case indexData := <-chElastic:
			indices = append(indices, indexData.Index)
			documentos = append(documentos, indexData.Data)
			if len(indices) == tamLote {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func despachaRedis() {

	//eu poderia acumular numa mapa de lists e, de tempos em tempos, flushar ao invés de enviar um por um.
	//não parece ser um problema agora não implementar isso, até porque eu não inventei nenhum caso ainda que exija.
	//mas vale a pena pensar sobre isso quando eu tiver casos em produção e ao monitorar a vazão e cpu e eventualmente
	//notar que não está muito bom.

	for indexData := range chRedis {
		b, err := json.Marshal(indexData.Data)
		if err != nil {
			log.Println(err)
			continue
		}
		if err := model.RedisCore.RPush(model.Ctx, indexData.Index, string(b)).Err(); err != nil {
			log.Println(err)
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

func extraiNginxRaw(n int) ([]NginxRaw, error) {
	lote := make([]NginxRaw, 0, n)
	result, err := model.RedisCore.BLPop(model.Ctx, 0, "nginx").Result()
	if err != nil {
		return nil, fmt.Errorf("extraiNginxRaw: %w", err)
	}
	var nr NginxRaw
	if err := json.Unmarshal([]byte(result[1]), &nr); err != nil {
		return nil, fmt.Errorf("extraiNginxRaw: %w", err)
	}
	lote = append(lote, nr)
	messages, err := model.RedisCore.LPopCount(model.Ctx, "nginx", n-1).Result()
	if errors.Is(err, redis.Nil) {
		return lote, nil
	}
	if err != nil {
		return nil, fmt.Errorf("extraiNginxRaw: %w", err)
	}
	for _, message := range messages {
		if err := json.Unmarshal([]byte(message), &nr); err != nil {
			return nil, fmt.Errorf("extraiNginxRaw: %w", err)
		}
		lote = append(lote, nr)
	}
	return lote, nil
}

func preprocessaNginx() {
	i := 0
	for {
		var ip gd.Ip
		lote, err := extraiNginxRaw(1)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		for _, nr := range lote {
			codUsuario, _ := strconv.Atoi(nr.XUserId)
			codPersonificador, _ := strconv.Atoi(nr.XUserId2)
			requestParts := strings.Split(nr.Request, " ")
			if len(requestParts) != 3 {
				log.Println("len(requestParts) != 3")
				continue
			}
			u, err := url.Parse(requestParts[1])
			if err != nil {
				log.Println(err)
				continue
			}
			i++
			ip, err = gd.FetchIp(nr.RemoteAddr)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 30)
			}
			n := Nginx{
				Time:              nr.Time,
				RemoteAddr:        nr.RemoteAddr,
				Hostname:          nr.Hostname,
				Host:              nr.Host,
				CodUsuario:        codUsuario,
				CodPersonificador: codPersonificador,
				Method:            requestParts[0],
				Endpoint:          u.Path,
				QueryString:       u.RawQuery,
				Status:            nr.Status,
				BodyBytesSent:     nr.BodyBytesSent,
				SessionId:         nr.XSessionId,
				HttpUserAgent:     nr.HttpUserAgent,
				HttpReferer:       nr.HttpReferer,
				GeoStatus:         ip.Status,
				CountryCode:       ip.CountryCode,
				Region:            ip.Region,
				City:              ip.City,
				Lat:               ip.Lat,
				Lon:               ip.Lon,
				As:                ip.As,
				Mobile:            ip.Mobile,
				Proxy:             ip.Proxy,
				Hosting:           ip.Hosting,
			}
			if i%1000 == 0 {
				fmt.Println(i)
			}
			chElastic <- IndexData{
				Index: "nginx",
				Data:  n,
			}
		}
	}
}
