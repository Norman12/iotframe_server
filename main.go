package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"gopkg.in/yaml.v2"

	"github.com/Norman12/iotframe_server/server"
)

const (
	DefaultTimeout time.Duration = 15 * time.Second
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadFile("configuration/configuration.yml"); err != nil {
		panic(err)
	}

	var c server.Configuration
	if err = yaml.Unmarshal(data, &c); err != nil {
		panic(err)
	}

	data = nil

	var (
		m = server.NewMedia(&c)
		s = server.NewServer(&c, m, logger)
		h = http.NewServeMux()
		p = &http.Server{
			Handler:      h,
			Addr:         *httpAddr,
			WriteTimeout: DefaultTimeout,
			ReadTimeout:  DefaultTimeout,
		}
	)

	h.Handle("/", s.NewRouter())

	errs := make(chan error, 2)
	go func() {
		logger.Info("app", zap.String("transport", "HTTP"), zap.String("addr", *httpAddr))
		errs <- p.ListenAndServe()
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Warn("app", zap.String("event", "terminating"), zap.Error(<-errs))
}
