/*
 * sunova - a player aid
 * Copyright (c) 2021  Michael D Henderson
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"github.com/mdhender/sunova/way"
	"log"
	"net"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC) // force logs to be UTC

	cfg := DefaultConfig()
	if err := cfg.Load(); err != nil {
		fmt.Printf("%+v\n", err)
	}

	if err := run(cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func run(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("missing configuration information")
	}
	s := &Server{
		debug:  cfg.Debug,
		DtFmt:  cfg.App.TimestampFormat,
		Router: way.NewRouter(),
	}
	s.Addr = net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
	s.IdleTimeout = cfg.Server.Timeout.Idle
	s.ReadTimeout = cfg.Server.Timeout.Read
	s.WriteTimeout = cfg.Server.Timeout.Write
	s.MaxHeaderBytes = cfg.Server.MaxHeaderBytes
	s.Handler = s.Router

	s.Routes()

	if cfg.Server.TLS.Serve {
		log.Printf("[main] serving TLS on %s\n", s.Addr)
		return s.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
	}
	log.Printf("[main] listening on %s\n", s.Addr)
	return s.ListenAndServe()
}

type Server struct {
	http.Server
	DtFmt  string // format string for timestamps in responses
	Router *way.Router
	debug  bool
}
