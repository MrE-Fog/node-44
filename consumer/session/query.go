/*
 * Copyright (C) 2020 The "MysteriumNetwork/node" Authors.
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

package session

import (
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/mysteriumnetwork/node/identity"
)

// Stats holds structure of aggregate session statistics.
type Stats struct {
	Count           int
	ConsumerCounts  map[identity.Identity]int
	SumDataSent     uint64
	SumDataReceived uint64
	SumDuration     time.Duration
	SumTokens       uint64
}

// NewQuery creates instance of new query.
func NewQuery() *Query {
	return &Query{
		where: make([]q.Matcher, 0),
		fetch: make([]q.Matcher, 0),
	}
}

// Query defines all flags for session filtering in session storage.
type Query struct {
	Sessions []History
	Stats    Stats

	where []q.Matcher
	fetch []q.Matcher
}

// FilterDirection filters fetched sessions by direction.
func (qr *Query) FilterDirection(direction string) *Query {
	qr.where = append(
		qr.where,
		matcher(func(session History) bool {
			return session.Direction == direction
		}),
	)
	return qr
}

// FetchSessions fetches list of sessions to Query.Sessions.
func (qr *Query) FetchSessions() *Query {
	qr.Sessions = make([]History, 0)

	qr.fetch = append(
		qr.fetch,
		matcher(func(session History) bool {
			qr.Sessions = append(qr.Sessions, session)
			return true
		}),
	)

	return qr
}

// FetchStats fetches sessions statistics to Query.Stats.
func (qr *Query) FetchStats() *Query {
	qr.Stats = Stats{
		ConsumerCounts: make(map[identity.Identity]int, 0),
	}

	qr.fetch = append(
		qr.fetch,
		matcher(func(session History) bool {
			qr.Stats.Count++

			if _, found := qr.Stats.ConsumerCounts[session.ConsumerID]; !found {
				qr.Stats.ConsumerCounts[session.ConsumerID] = 1
			} else {
				qr.Stats.ConsumerCounts[session.ConsumerID]++
			}

			qr.Stats.SumDataReceived += session.DataReceived
			qr.Stats.SumDataSent += session.DataSent
			qr.Stats.SumDuration += session.GetDuration()
			qr.Stats.SumTokens += session.Tokens

			return true
		}),
	)

	return qr
}

func (qr *Query) toQuery(node storm.Node) storm.Query {
	return node.Select(
		q.And(qr.where...),
		q.And(qr.fetch...),
	)
}

type matcher func(History) bool

func (m matcher) Match(i interface{}) (bool, error) {
	return m(i.(History)), nil
}
