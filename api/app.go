// deis-logger
// https://github.com/topfreegames/deis-logger
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"io"
	"net"
	"net/http"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/garyburd/redigo/redis"
	raven "github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	newrelic "github.com/newrelic/go-agent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/deis-logger/metadata"
)

//App is our API application
type App struct {
	Address             string
	Config              *viper.Viper
	ElasticSearchClient *elastic.Client
	Logger              logrus.FieldLogger
	NewRelic            newrelic.Application
	Router              *mux.Router
	Server              *http.Server
	Redis               *redis.Pool
}

// NewApp Creates App instance
func NewApp(
	host string,
	port int,
	config *viper.Viper,
	logger logrus.FieldLogger,
	esClientOrNil *elastic.Client,
	redisPoolOrNil *redis.Pool,
) (*App, error) {
	a := &App{
		Config:              config,
		Address:             fmt.Sprintf("%s:%d", host, port),
		Logger:              logger,
		ElasticSearchClient: esClientOrNil,
		Redis:               redisPoolOrNil,
	}
	err := a.configureApp(esClientOrNil)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) loadConfigurationDefaults() {
	a.Config.SetDefault("elasticsearch.url", "http://localhost:9200")
	a.Config.SetDefault("redis.url", "redis://localhost:6379/0")
}

func (a *App) configureLogger() {
	a.Logger = a.Logger.WithFields(logrus.Fields{
		"source":    "deis-logger",
		"operation": "initializeApp",
		"version":   metadata.Version,
	})
}

func (a *App) configureElasticSearch(esClient *elastic.Client) error {
	esURL := a.Config.GetString("elasticsearch.url")
	esSniff := a.Config.GetBool("elasticsearch.sniff")

	l := a.Logger.WithFields(logrus.Fields{
		"sniff": esSniff,
		"esURL": esURL,
	})

	l.Info("Connecting to Elasticsearch...")

	es, err := elastic.NewClient(
		elastic.SetURL(esURL),
		elastic.SetSniff(esSniff),
	)

	a.ElasticSearchClient = es
	return err
}

func (a *App) configureApp(esClientOrNil *elastic.Client) error {
	a.loadConfigurationDefaults()
	a.configureLogger()
	err := a.configureElasticSearch(esClientOrNil)
	if err != nil {
		return err
	}

	a.configureRedis()

	err = a.configureNewRelic()
	if err != nil {
		return err
	}

	a.configureSentry()

	a.configureServer()

	return nil
}

//ListenAndServe requests
func (a *App) ListenAndServe() (io.Closer, error) {
	listener, err := net.Listen("tcp", a.Address)
	if err != nil {
		return nil, err
	}

	err = a.Server.Serve(listener)
	if err != nil {
		listener.Close()
		return nil, err
	}

	return listener, nil
}

func (a *App) configureSentry() {
	l := a.Logger.WithFields(logrus.Fields{
		"operation": "configureSentry",
	})
	sentryURL := a.Config.GetString("sentry.url")
	l.Debug("Configuring sentry...")
	raven.SetDSN(sentryURL)
	raven.SetRelease(metadata.Version)
}

func (a *App) configureRedis() {
	redisURL := a.Config.GetString("redis.url")
	a.Redis = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL)
		},
	}
}

func (a *App) configureNewRelic() error {
	appName := a.Config.GetString("newrelic.app")
	key := a.Config.GetString("newrelic.key")

	l := a.Logger.WithFields(logrus.Fields{
		"appName":   appName,
		"operation": "configureNewRelic",
	})

	if key == "" {
		l.Warning("New Relic key not found. No data will be sent to New Relic.")
		return nil
	}

	l.Debug("Configuring new relic...")
	config := newrelic.NewConfig(appName, key)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		l.WithError(err).Error("Failed to configure new relic.")
		return err
	}

	l.WithFields(logrus.Fields{
		"key": key,
	}).Info("New Relic configured successfully.")
	a.NewRelic = app
	return nil
}

func (a *App) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/healthz", NewHealthcheckHandler(a)).Methods("GET")
	r.Handle("/healthz/", NewHealthcheckHandler(a)).Methods("GET")
	r.Handle("/logs/{app}", NewLogHandler(a)).Methods("GET")
	r.Handle("/logs/{app}/", NewLogHandler(a)).Methods("GET")
	r.Handle("/logs/{app}/tail", NewLogTailHandler(a)).Methods("GET")
	r.Handle("/logs/{app}/tail/", NewLogTailHandler(a)).Methods("GET")
	return r
}

func (a *App) configureServer() {
	a.Router = a.getRouter()
	a.Server = &http.Server{Addr: a.Address, Handler: a.Router}
}
