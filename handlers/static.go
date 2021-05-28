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

package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Static returns a new static file handler.
func Static(prefix string, root string, spaRouting, debug bool) http.Handler {
	log.Println("[static] initializing")
	defer log.Println("[static] initialized")

	log.Printf("[static] strip: %q\n", prefix)
	log.Printf("[static]  root: %q\n", root)

	var indexFile string
	if spaRouting {
		indexFile = filepath.Join(root, "index.html")
		log.Printf("[static] index: %q\n", indexFile)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		if debug {
			log.Printf("[static] %q\n", file)
		}

		stat, err := os.Stat(file)
		if err != nil {
			if spaRouting {
				// try serving index file for SPA routing instead
				if stat, err := os.Stat(indexFile); err == nil {
					if rdr, err := os.Open(indexFile); err == nil {
						defer func(r io.ReadCloser) {
							_ = r.Close()
						}(rdr)
						http.ServeContent(w, r, file, stat.ModTime(), rdr)
						return
					}
				}
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// we never want to give a directory listing, so change raw directory request to fetch the index.html instead.
		if stat.IsDir() {
			file = filepath.Join(file, "index.html")
			stat, err = os.Stat(file)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}

		// only serve regular files (this avoids serving a directory named index.html)
		if !stat.Mode().IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer func(r io.ReadCloser) {
			_ = r.Close()
		}(rdr)

		http.ServeContent(w, r, file, stat.ModTime(), rdr)
	})
}
