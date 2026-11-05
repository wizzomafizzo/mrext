// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rs/cors"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// originPolicy decides which browser origins may call the API. Requests
// without browser origin headers, such as native apps and curl, are never
// affected. Browsers get same-origin access, access from other Remote web UIs
// on the local network, and access from origins listed in allowed_origins.
type originPolicy struct {
	trusted  map[string]bool
	allowAll bool
}

// newOriginPolicy builds a policy from allowed_origins entries. An entry of
// "*" allows every origin. Invalid entries are skipped and returned as errors.
func newOriginPolicy(entries []string) (*originPolicy, []error) {
	p := &originPolicy{trusted: make(map[string]bool)}
	var errs []error
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == "*" {
			p.allowAll = true
			continue
		}
		origin, err := normalizeOrigin(entry)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		p.trusted[origin] = true
	}
	return p, errs
}

func normalizeOrigin(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSuffix(raw, "/"))
	if err != nil {
		return "", fmt.Errorf("invalid allowed origin %q: %w", raw, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" ||
		u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("invalid allowed origin %q: use scheme://host[:port]", raw)
	}
	return strings.ToLower(u.Scheme + "://" + u.Host), nil
}

// originAllowed reports whether a cross-origin page may use the API.
func (p *originPolicy) originAllowed(origin string) bool {
	if p.allowAll {
		return true
	}
	normalized, err := normalizeOrigin(origin)
	if err != nil {
		return false
	}
	if p.trusted[normalized] {
		return true
	}
	return isLocalRemoteOrigin(normalized)
}

// isLocalRemoteOrigin matches the web UI served by another Remote on the local
// network, which can be pointed at this MiSTer. Public sites cannot obtain an
// origin with a private IP literal, a .local name or a single-label name.
func isLocalRemoteOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" || u.Port() != strconv.Itoa(appPort) {
		return false
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
	}
	host = strings.TrimSuffix(host, ".")
	return host != "" && (!strings.Contains(host, ".") || strings.HasSuffix(host, ".local"))
}

// requestAllowed reports whether a request comes from an allowed browser
// context. It follows the same rules as net/http.CrossOriginProtection, with
// originAllowed in place of a fixed trusted origin list.
func (p *originPolicy) requestAllowed(r *http.Request) bool {
	if p.allowAll {
		return true
	}
	origin := r.Header.Get("Origin")
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "none":
		return true
	case "":
		if origin == "" {
			return true
		}
		if u, err := url.Parse(origin); err == nil && strings.EqualFold(u.Host, r.Host) {
			return true
		}
		return p.originAllowed(origin)
	default:
		return origin != "" && p.originAllowed(origin)
	}
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func forbidOrigin(w http.ResponseWriter) {
	http.Error(w, "cross-origin request not allowed; see allowed_origins in remote.ini",
		http.StatusForbidden)
}

// Handler applies CORS headers and rejects state-changing requests from
// browser origins that are not allowed.
func (p *originPolicy) Handler(next http.Handler) http.Handler {
	guarded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isSafeMethod(r.Method) && !p.requestAllowed(r) {
			forbidOrigin(w)
			return
		}
		next.ServeHTTP(w, r)
	})
	opts := cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With", mister.IniFilenameHeader},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
	}
	if p.allowAll {
		opts.AllowedOrigins = []string{"*"}
	} else {
		opts.AllowOriginFunc = p.originAllowed
	}
	return cors.New(opts).Handler(guarded)
}

// LaunchGuard protects GET launch links. A visible top-level navigation, such
// as a scanned QR code or a clicked link, is allowed. Hidden cross-site loads
// through images, frames or scripts are rejected.
func (p *originPolicy) LaunchGuard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topLevel := r.Header.Get("Sec-Fetch-Mode") == "navigate" &&
			r.Header.Get("Sec-Fetch-Dest") == "document"
		if !topLevel && !p.requestAllowed(r) {
			forbidOrigin(w)
			return
		}
		next(w, r)
	}
}
