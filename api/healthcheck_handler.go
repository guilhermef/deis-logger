// deis-logger
// https://github.com/topfreegames/deis-logger
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package api

import "net/http"

//HealthcheckHandler handler
type HealthcheckHandler struct {
	App *App
}

// NewHealthcheckHandler creates a new healthcheck handler
func NewHealthcheckHandler(a *App) *HealthcheckHandler {
	m := &HealthcheckHandler{App: a}
	return m
}

//ServeHTTP method
func (h *HealthcheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
