// deis-logger
// https://github.com/topfreegames/deis-logger
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/gorilla/mux"
)

// LogMessage struct
type LogMessage struct {
	Log        string `json:"log"`
	Kubernetes struct {
		Labels struct {
			Type string `json:"type"`
		} `json:"labels"`
	} `json:"kubernetes"`
}

//LogHandler handler
type LogHandler struct {
	App *App
}

// NewLogHandler creates a new healthcheck handler
func NewLogHandler(a *App) *LogHandler {
	m := &LogHandler{App: a}
	return m
}

//ServeHTTP method
func (h *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logLines := 100
	logLinesParam := r.URL.Query().Get("log_lines")
	if logLinesParam != "" {
		res, _ := strconv.ParseInt(logLinesParam, 10, 64)
		logLines = int(res)
	}
	ctx := context.Background()
	vars := mux.Vars(r)
	fetchSourceContext := elastic.NewFetchSourceContext(true).Include("log", "kubernetes.labels.type")
	searchResult, _ := h.App.ElasticSearchClient.
		Search().
		Index("k8s-"+vars["app"]+"-stash-*").
		Sort("@timestamp", false).
		FetchSourceContext(fetchSourceContext).
		Size(logLines).
		Do(ctx)

	for i := range searchResult.Hits.Hits {
		var l LogMessage
		hit := searchResult.Hits.Hits[len(searchResult.Hits.Hits)-1-i]
		err := json.Unmarshal(*hit.Source, &l)
		if err != nil {
			h.App.Logger.Error(err)
		}
		// io.Copy(w, fmt.Fprintf("[%s]%s", l.Kubernetes.Labels.Type, l.Log))
		fmt.Fprintf(w, "[%s]%s", l.Kubernetes.Labels.Type, l.Log)
	}
}
