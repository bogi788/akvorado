// SPDX-FileCopyrightText: 2022 Free Mobile
// SPDX-License-Identifier: AGPL-3.0-only

package routing

import (
	"net/http"
	"net/netip"

	"akvorado/common/helpers"

	"github.com/gin-gonic/gin"
)

type routeParameters struct {
	ip    netip.Addr `form:"ip"`
	nh    netip.Addr `form:"next_hop"`
	agent netip.Addr `form:"agent"`
}

// RoutesHTTPHandler looks up a route and sends it as JSON
func (c *Component) RoutesHTTPHandler(gc *gin.Context) {
	var params routeParameters
	if err := gc.ShouldBindQuery(&params); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"message": helpers.Capitalize(err.Error())})
		return
	}

	gc.JSON(200, "test")
}
