package main

import (
	"context"
	crypto "crypto/rand"
	"encoding/xml"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/remizovm/geonames"
	"github.com/remizovm/geonames/models"
	"googlemaps.github.io/maps"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var pool *pgxpool.Pool
var geocoder *maps.Client
var geonameCities map[int]*models.Feature
var dataDog *statsd.Client
var config *Config

func checkError(err error) {
	if err != nil {
		log.Fatalf("Demae Dominos server has encountered an error! Reason: %v\n", err)
	}
}

func main() {
	rawConfig, err := os.ReadFile("./config.xml")
	checkError(err)

	config = &Config{}
	err = xml.Unmarshal(rawConfig, config)
	checkError(err)

	// Before we do anything, init Sentry to capture all errors.
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              config.SentryDSN,
		Debug:            true,
		TracesSampleRate: 1.0,
	})
	checkError(err)
	defer sentry.Flush(2 * time.Second)

	// Initialize Datadog
	tracer.Start(tracer.WithService("demae-dominos"), tracer.WithAgentAddr("127.0.0.1:8126"))
	defer tracer.Stop()

	err = profiler.Start(profiler.WithService("demae-dominos"))
	checkError(err)
	defer profiler.Stop()

	dataDog, err = statsd.New("127.0.0.1:8125")
	checkError(err)

	// Initialize database
	dbString := fmt.Sprintf("postgres://%s:%s@%s/%s", config.SQLUser, config.SQLPass, config.SQLAddress, config.SQLDB)
	dbConf, err := pgxpool.ParseConfig(dbString)
	checkError(err)
	pool, err = pgxpool.ConnectConfig(context.Background(), dbConf)
	checkError(err)

	// Ensure this Postgresql connection is valid.
	defer pool.Close()

	// Initialize Google Maps API
	geocoder, err = maps.NewClient(maps.WithAPIKey(config.GoogleMapsAPIKey))

	// Initialize Geonames cities array.
	client := geonames.Client{}
	geonameCities, err = client.Cities15000()
	checkError(err)

	// Seed random number generator
	seed, err := crypto.Int(crypto.Reader, big.NewInt(math.MaxInt64))
	checkError(err)
	rand.Seed(seed.Int64())

	// Finally, initialize the HTTP server
	fmt.Printf("Starting HTTP connection (%s)...\nNot using the usual port for HTTP?\nBe sure to use a proxy, otherwise the Wii can't connect!\n", config.Address)
	r := NewRoute()
	nwapi := r.HandleGroup("nwapi.php")
	{
		nwapi.NormalResponse("webApi_document_template", documentTemplate)
		nwapi.NormalResponse("webApi_area_list", areaList)
		nwapi.MultipleRootNodes("webApi_category_list", categoryList)
		nwapi.NormalResponse("webApi_area_shopinfo", func(r *Response) {})
		nwapi.NormalResponse("webApi_shop_list", categoryList)
		nwapi.MultipleRootNodes("webApi_shop_one", shopOne)
		nwapi.MultipleRootNodes("webApi_menu_list", menuList)
		nwapi.MultipleRootNodes("webApi_item_list", itemList)
		nwapi.MultipleRootNodes("webApi_item_one", itemOne)
		nwapi.MultipleRootNodes("webApi_Authkey", authKey)
		nwapi.MultipleRootNodes("webApi_basket_delete", basketDelete)
		nwapi.MultipleRootNodes("webApi_basket_reset", basketReset)
		nwapi.MultipleRootNodes("webApi_basket_add", basketAdd)
		nwapi.MultipleRootNodes("webApi_basket_list", basketList)
		nwapi.MultipleRootNodes("webApi_validate_condition", func(r *Response) {})
		nwapi.NormalResponse("webApi_order_done", orderDone)
		nwapi.NormalResponse("webApi_inquiry_done", inquiryDone)
	}

	log.Fatal(http.ListenAndServe(config.Address, r.Handle()))
}
