package gd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	IpTtl = 86400 * 7
)

var (
	mips     = make(map[string]Ip)
	NextFech time.Time
)

type Ip struct {
	Id          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Ip          string        `json:"ip" bson:"ip"`
	Request     time.Time     `json:"request" bson:"request"`
	Response    time.Time     `json:"response" bson:"response"`
	Status      string        `json:"status" bson:"status"`
	Message     string        `json:"message" bson:"message"`
	Country     string        `json:"country" bson:"country"`
	CountryCode string        `json:"countryCode" bson:"countryCode"`
	Region      string        `json:"region" bson:"region"`
	RegionName  string        `json:"regionName" bson:"regionName"`
	City        string        `json:"city" bson:"city"`
	Zip         string        `json:"zip" bson:"zip"`
	Lat         float64       `json:"lat" bson:"lat"`
	Lon         float64       `json:"lon" bson:"lon"`
	Timezone    string        `json:"timezone" bson:"timezone"`
	Isp         string        `json:"isp" bson:"isp"`
	Org         string        `json:"org" bson:"org"`
	As          string        `json:"as" bson:"as"`
	Mobile      bool          `json:"mobile" bson:"mobile"`
	Proxy       bool          `json:"proxy" bson:"proxy"`
	Hosting     bool          `json:"hosting" bson:"hosting"`
	Query       string        `json:"query" bson:"query"`
}

// FetchIp retorna dados de geolocalização sincronamente da forma mais rápida disponível, nessa ordem:
//
// 1. mapa de ips
// 2. banco de dados, collection gd.ips, em MongoGo
// 3. ip-api.com
//
// Acessa graciosamente ip-api.com, com no máximo uma consulta por segundo, atrasando a resposta se for preciso.
// Útil para quem precisa esperar pela resposta, bem-sucedida ou não, antes de fazer a próxima pesquisa.
func FetchIp(ip string) (Ip, error) {
	var (
		info Ip
		ok   bool
	)
	info, ok = mips[ip]
	if ok {
		return info, nil
	}
	err := db.Collection("ips").FindOne(nil, bson.M{"ip": ip}).Decode(&info)
	if err == nil && time.Now().Sub(info.Request).Seconds() < IpTtl {
		info.Request = info.Request.Local()
		info.Response = info.Response.Local()
		mips[ip] = info
		return info, nil
	}
	time.Sleep(NextFech.Sub(time.Now()))
	info, err = resolveIp(ip)
	if err != nil {
		return info, fmt.Errorf("fetchip: %w", err)
	}
	info.Ip = ip
	info.Request = time.Now()
	info.Response = time.Now()
	opts := options.Replace().SetUpsert(true)
	if _, err := db.Collection("ips").ReplaceOne(nil, bson.M{"ip": ip}, info, opts); err != nil {
		return info, fmt.Errorf("fetchip: %w", err)
	}
	mips[ip] = info
	NextFech = time.Now().Add(time.Second * 2)
	return info, nil
}

func QueryIp(ip string) (Ip, error) {
	ip = strings.TrimSpace(ip)
	var info Ip
	err := db.Collection("ips").FindOne(nil, bson.M{"ip": ip}).Decode(&info)
	if err == nil && time.Now().Sub(info.Request).Seconds() < IpTtl {
		info.Request = info.Request.Local()
		info.Response = info.Response.Local()
		return info, nil
	}
	ips, err := RequestedIps()
	if err != nil {
		return Ip{}, fmt.Errorf("query ip: %v", err)
	}
	if len(ips) == 0 {
		info, err := resolveIp(ip)
		if err != nil {
			return Ip{}, fmt.Errorf("query ip: %v", err)
		}
		info.Ip = ip
		info.Request = time.Now()
		info.Response = time.Now()
		if err := saveIp(info); err != nil {
			return Ip{}, fmt.Errorf("query ip: %v", err)
		}
		return info, nil
	}
	info.Ip = ip
	info.Request = time.Now()
	info.Status = "requested"
	opts := options.Replace().SetUpsert(true)
	if _, err := db.Collection("ips").ReplaceOne(nil, bson.M{"ip": ip}, info, opts); err != nil {
		return info, fmt.Errorf("query ip: %v", err)
	}
	return info, nil
}

func resolveIp(ip string) (Ip, error) {
	ip = strings.TrimSpace(ip)
	var info Ip
	res, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%v?fields=17035263", ip))
	if err != nil {
		return Ip{}, fmt.Errorf("resolve ip: %v", err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return Ip{}, fmt.Errorf("resolve ip: %v", err)
	}
	if err = json.Unmarshal(b, &info); err != nil {
		return Ip{}, fmt.Errorf("resolve ip: %v", err)
	}
	return info, nil
}

func RequestedIps() ([]Ip, error) {
	opts := options.Find().SetSort(bson.D{{"request", -1}})
	cur, err := db.Collection("ips").Find(nil, bson.M{"status": "requested"}, opts)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Minute)
	}
	ips := make([]Ip, 0)
	if err := cur.All(nil, &ips); err != nil {
		return ips, fmt.Errorf("requested ips: %w", err)
	}
	return ips, nil
}

func saveIp(info Ip) error {
	opts := options.Replace().SetUpsert(true)
	if _, err := db.Collection("ips").ReplaceOne(nil, bson.M{"ip": info.Ip}, info, opts); err != nil {
		return fmt.Errorf("save ip: %v", err)
	}
	return nil
}

func ResolveIps() {
	for {
		ips, err := RequestedIps()
		if err != nil {
			log.Println(err)
			time.Sleep(time.Minute)
			continue
		}
		for _, ip := range ips {
			info, err := resolveIp(ip.Ip)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second)
				continue
			}
			info.Ip = ip.Ip
			info.Request = ip.Request
			info.Response = time.Now()
			if err := saveIp(info); err != nil {
				log.Println(err)
				time.Sleep(time.Second)
				continue
			}
			time.Sleep(time.Second)
		}
		time.Sleep(5 * time.Minute)
	}
}
