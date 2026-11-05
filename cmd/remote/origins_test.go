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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type originRequest struct {
	headers map[string]string
	method  string
	path    string
	host    string
}

func newOriginRequest(req originRequest) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), req.method, "http://"+req.host+req.path, http.NoBody)
	r.Host = req.host
	for k, v := range req.headers {
		r.Header.Set(k, v)
	}
	return r
}

func serveThrough(h http.Handler, r *http.Request) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec.Code
}

func TestOriginPolicyStateChangingRequests(t *testing.T) {
	t.Parallel()

	policy, errs := newOriginPolicy([]string{"https://Dashboard.example/", "not a url", "ftp://x.example"})
	if len(errs) != 2 {
		t.Fatalf("errs = %v, want 2 invalid entries", errs)
	}
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := policy.Handler(ok)

	tests := []struct {
		headers map[string]string
		name    string
		want    int
	}{
		{name: "native client", want: http.StatusOK},
		{name: "same origin header", headers: map[string]string{"Origin": "http://mister:8182"}, want: http.StatusOK},
		{
			name:    "same origin fetch metadata",
			headers: map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "http://mister:8182"},
			want:    http.StatusOK,
		},
		{
			name:    "hostile page",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://evil.example"},
			want:    http.StatusForbidden,
		},
		{
			name:    "hostile page old browser",
			headers: map[string]string{"Origin": "http://evil.example"},
			want:    http.StatusForbidden,
		},
		{
			name:    "hostile page on remote port",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://evil.example:8182"},
			want:    http.StatusForbidden,
		},
		{
			name:    "cross site without origin",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site"},
			want:    http.StatusForbidden,
		},
		{
			name:    "configured origin",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://dashboard.example"},
			want:    http.StatusOK,
		},
		{
			name:    "other Remote by IP",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://192.168.1.20:8182"},
			want:    http.StatusOK,
		},
		{
			name:    "other Remote by mDNS name",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://mister2.local:8182"},
			want:    http.StatusOK,
		},
		{
			name:    "other Remote by short name",
			headers: map[string]string{"Sec-Fetch-Site": "same-site", "Origin": "http://mister2:8182"},
			want:    http.StatusOK,
		},
		{
			name:    "private IP on another port",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://192.168.1.20:8080"},
			want:    http.StatusForbidden,
		},
		{
			name:    "public IP on remote port",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://203.0.113.5:8182"},
			want:    http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := newOriginRequest(originRequest{
				method: http.MethodPost, path: "/api/settings/system/reboot", host: "mister:8182", headers: tt.headers,
			})
			if got := serveThrough(h, r); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOriginPolicyAllowsReads(t *testing.T) {
	t.Parallel()

	policy, _ := newOriginPolicy(nil)
	h := policy.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r := newOriginRequest(originRequest{
		method: http.MethodGet, path: "/api/sysinfo", host: "mister:8182",
		headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://evil.example"},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want none", got)
	}
}

func TestOriginPolicyCORSHeaders(t *testing.T) {
	t.Parallel()

	policy, _ := newOriginPolicy([]string{"https://dashboard.example"})
	h := policy.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		origin string
		want   string
	}{
		{origin: "http://evil.example", want: ""},
		{origin: "https://dashboard.example", want: "https://dashboard.example"},
		{origin: "http://10.0.0.4:8182", want: "http://10.0.0.4:8182"},
	}
	for _, tt := range tests {
		r := newOriginRequest(originRequest{
			method: http.MethodOptions, path: "/api/settings/system/reboot", host: "mister:8182",
			headers: map[string]string{"Origin": tt.origin, "Access-Control-Request-Method": "POST"},
		})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.want {
			t.Fatalf("%s: Access-Control-Allow-Origin = %q, want %q", tt.origin, got, tt.want)
		}
	}
}

func TestOriginPolicyWildcard(t *testing.T) {
	t.Parallel()

	policy, errs := newOriginPolicy([]string{" * "})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
	h := policy.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r := newOriginRequest(originRequest{
		method: http.MethodPost, path: "/api/launch", host: "mister:8182",
		headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "http://evil.example"},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("status = %d, ACAO = %q", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestOriginPolicyLaunchGuard(t *testing.T) {
	t.Parallel()

	policy, _ := newOriginPolicy(nil)
	h := policy.LaunchGuard(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	tests := []struct {
		headers map[string]string
		name    string
		want    int
	}{
		{name: "native client", want: http.StatusOK},
		{
			name: "QR code or typed URL",
			headers: map[string]string{
				"Sec-Fetch-Site": "none", "Sec-Fetch-Mode": "navigate", "Sec-Fetch-Dest": "document",
			},
			want: http.StatusOK,
		},
		{
			name: "link clicked on another site",
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "navigate", "Sec-Fetch-Dest": "document",
			},
			want: http.StatusOK,
		},
		{
			name: "hidden image",
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "no-cors", "Sec-Fetch-Dest": "image",
			},
			want: http.StatusForbidden,
		},
		{
			name: "hidden frame",
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "navigate", "Sec-Fetch-Dest": "iframe",
			},
			want: http.StatusForbidden,
		},
		{
			name: "script fetch",
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "cors", "Sec-Fetch-Dest": "empty",
				"Origin": "http://evil.example",
			},
			want: http.StatusForbidden,
		},
		{
			name: "web UI on this MiSTer",
			headers: map[string]string{
				"Sec-Fetch-Site": "same-origin", "Sec-Fetch-Mode": "cors", "Sec-Fetch-Dest": "empty",
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := newOriginRequest(originRequest{
				method: http.MethodGet, path: "/api/l/bWVudS5yYmY=", host: "mister:8182", headers: tt.headers,
			})
			rec := httptest.NewRecorder()
			h(rec, r)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestOriginPolicyWebSocket(t *testing.T) {
	t.Parallel()

	policy, _ := newOriginPolicy(nil)
	tests := []struct {
		origin string
		want   bool
	}{
		{origin: "", want: true},
		{origin: "http://mister:8182", want: true},
		{origin: "http://MISTER:8182", want: true},
		{origin: "http://evil.example", want: false},
	}
	for _, tt := range tests {
		r := newOriginRequest(originRequest{
			method: http.MethodGet, path: "/api/ws", host: "mister:8182",
			headers: map[string]string{"Origin": tt.origin},
		})
		if tt.origin == "" {
			r.Header.Del("Origin")
		}
		if got := policy.requestAllowed(r); got != tt.want {
			t.Fatalf("%q: allowed = %v, want %v", tt.origin, got, tt.want)
		}
	}
}
