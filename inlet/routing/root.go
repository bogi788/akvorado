// SPDX-FileCopyrightText: 2023 Free Mobile
// SPDX-License-Identifier: AGPL-3.0-only

// Package routing fetches routing-related data (AS numbers, AS paths,
// communities). It is modular and accepts several kind of providers (including
// BMP).
package routing

import (
	"context"
	"net/netip"
	"time"

	"github.com/benbjohnson/clock"

	"akvorado/common/daemon"
	"akvorado/common/httpserver"
	"akvorado/common/reporter"
	"akvorado/inlet/routing/provider"
)

// Component represents the metadata compomenent.
type Component struct {
	r         *reporter.Reporter
	d         *Dependencies
	provider  provider.Provider
	metrics   metrics
	config    Configuration
	errLogger reporter.Logger
}

// Dependencies define the dependencies of the metadata component.
type Dependencies struct {
	Daemon daemon.Component
	Clock  clock.Clock
	HTTP   *httpserver.Component
}

// New creates a new metadata component.
func New(r *reporter.Reporter, configuration Configuration, dependencies Dependencies) (*Component, error) {
	c := Component{
		r:         r,
		d:         &dependencies,
		config:    configuration,
		errLogger: r.Sample(reporter.BurstSampler(time.Minute, 3)),
	}
	c.initMetrics()
	// Initialize the provider
	selectedProvider, err := configuration.Provider.Config.New(r, provider.Dependencies{dependencies.Daemon, dependencies.Clock})
	if err != nil {
		return nil, err
	}
	c.provider = selectedProvider

	return &c, nil
}

// Start starts the routing component.
func (c *Component) Start() error {
	c.r.Info().Msg("starting routing component")
	if starterP, ok := c.provider.(starter); ok {
		if err := starterP.Start(); err != nil {
			return err
		}
	}

	c.d.HTTP.GinRouter.GET("/api/v0/inlet/route", c.RouteHTTPHandler)
	return nil
}

// Stop stops the routing component
func (c *Component) Stop() error {
	c.r.Info().Msg("stopping routing component")
	if stopperP, ok := c.provider.(stopper); ok {
		if err := stopperP.Stop(); err != nil {
			return err
		}
	}
	return nil
}

type starter interface {
	Start() error
}
type stopper interface {
	Stop() error
}

// Lookup uses the selected provider to get an answer.
func (c *Component) Lookup(ctx context.Context, ip netip.Addr, nh netip.Addr, agent netip.Addr) provider.LookupResult {
	c.metrics.routingLookups.Inc()
	result, err := c.provider.Lookup(ctx, ip, nh, agent)
	if err != nil {
		c.metrics.routingLookupsFailed.Inc()
		c.errLogger.Err(err).Msgf("routing: error while looking up %s at %s", ip.String(), agent.String())
	}
	return result
}
