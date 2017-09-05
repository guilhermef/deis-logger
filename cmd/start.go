// deis-logger
// https://github.com/topfreegames/deis-logger
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package cmd

import (

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/topfreegames/deis-logger/api"
)

var bind string
var port int
var incluster bool
var context string
var kubeconfig string

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts deis-logger",
	Long:  `starts deis-logger`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := logrus.InfoLevel
		switch Verbose {
		case 0:
			ll = logrus.InfoLevel
		case 1:
			ll = logrus.WarnLevel
		case 3:
			ll = logrus.DebugLevel
		}

		var log = logrus.New()
		if json {
			log.Formatter = new(logrus.JSONFormatter)
		}
		log.Level = ll

		cmdL := log.WithFields(logrus.Fields{
			"source":    "startCmd",
			"operation": "Run",
			"bind":      bind,
			"port":      port,
		})

		cmdL.Info("starting deis-logger")

		app, err := api.NewApp(bind, port, config, log, nil, nil)
		if err != nil {
			cmdL.Fatal(err)
		}

		app.ListenAndServe()

	},
}

func init() {
	startCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "bind address")
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "bind port")
	RootCmd.AddCommand(startCmd)
}
