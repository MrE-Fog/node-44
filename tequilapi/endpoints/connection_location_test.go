/*
 * Copyright (C) 2019 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package endpoints

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/mysteriumnetwork/node/core/connection"
	"github.com/stretchr/testify/assert"
)

func TestAddRoutesForConnectionLocationAddsRoutes(t *testing.T) {
	router := httprouter.New()

	AddRoutesForConnectionLocation(router, &mockConnectionManager{
		onStatusReturn: connection.Status{
			State: connection.Connected,
		},
	}, &locationResolverMock{
		ip: "1.2.3.4",
	})

	tests := []struct {
		method         string
		path           string
		body           string
		expectedStatus int
		expectedJSON   string
	}{
		{
			http.MethodGet, "/connection/location", "",
			http.StatusOK,
			`{
				"asn": 62179,
				"city": "Vilnius",
				"continent": "EU",
				"country": "LT",
				"ip": "1.2.3.4",
				"isp": "Telia Lietuva, AB",
				"node_type": "residential"
			}`,
		},
	}

	for _, test := range tests {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		router.ServeHTTP(resp, req)
		assert.Equal(t, test.expectedStatus, resp.Code)
		if test.expectedJSON != "" {
			assert.JSONEq(t, test.expectedJSON, resp.Body.String())
		} else {
			assert.Equal(t, "", resp.Body.String())
		}
	}
}

func TestAddRoutesForConnectionLocationFailOnDisconnectedState(t *testing.T) {
	router := httprouter.New()

	AddRoutesForConnectionLocation(router, &mockConnectionManager{}, &locationResolverMock{ip: "1.2.3.4"})

	tests := []struct {
		method         string
		path           string
		body           string
		expectedStatus int
		expectedJSON   string
	}{
		{
			http.MethodGet, "/connection/location", "",
			http.StatusServiceUnavailable,
			`{"message":"Connection is not connected"}`,
		},
	}

	for _, test := range tests {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		router.ServeHTTP(resp, req)
		assert.Equal(t, test.expectedStatus, resp.Code)
		if test.expectedJSON != "" {
			assert.JSONEq(t, test.expectedJSON, resp.Body.String())
		} else {
			assert.Equal(t, "", resp.Body.String())
		}
	}
}
