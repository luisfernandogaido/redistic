package gd

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	uriPro = "mongodb://gaido:1000sonhosreais@167.99.55.99:27017/?authSource=admin"
	uriLoc = "mongodb://localhost:27017"

	lenHash = 7
)

var (
	client   *mongo.Client
	db       *mongo.Database
	hostName string
)

func init() {
	var (
		uri string
		err error
	)
	hostName, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	if hostName == "NOTE-GAIDO" || hostName == "DESK-GAIDO" {
		//uri = uriLoc
		uri = uriPro
	} else {
		uri = uriPro
	}
	client, err = mongo.Connect(nil, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	if err := client.Ping(nil, readpref.Primary()); err != nil {
		panic(err)
	}
	db = client.Database("gd")
}

func textSearch(s string) string {
	termos := strings.Split(s, " ")
	termos2 := make([]string, 0)
	for _, t := range termos {
		if t == "" {
			continue
		}
		termo := fmt.Sprintf("%q", t)
		if strings.HasPrefix(t, "-") && len(t) > 1 {
			termo = fmt.Sprintf("-%q", t[1:])
		}
		termos2 = append(termos2, termo)
	}
	return strings.Join(termos2, " ")
}

func generateHash() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	var hash strings.Builder
	for i := 0; i < lenHash; i++ {
		i := r.Intn(64)
		hash.WriteString(chars[i : i+1])
	}
	return hash.String()
}

type Document map[string]any
