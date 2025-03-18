// SPDX-FileCopyrightText: 2022 Free Mobile
// SPDX-License-Identifier: AGPL-3.0-only

package routing

import (
	"fmt"
	"net/http"
	"net/netip"

	"akvorado/common/helpers"

	"github.com/gin-gonic/gin"
)

type routeParameters struct {
	IP string `form:"ip"`
}

// RouteHTTPHandler looks up a route and sends it as JSON
func (c *Component) RouteHTTPHandler(gc *gin.Context) {
	var params routeParameters
	if err := gc.ShouldBindQuery(&params); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"message": helpers.Capitalize(err.Error())})
		return
	}

	ip, err := netip.ParseAddr(params.IP)
	if err != nil {
		gc.JSON(500, gin.H{"error": fmt.Sprintf("failed to parse ip %s: %v", params.IP, err)})
		return
	}
	if ip.Is4() {
		ip = netip.AddrFrom16(ip.As16())
	}

	result, err := c.provider.LookupRoutes(ip)
	if err != nil {
		gc.JSON(404, gin.H{"error": fmt.Sprintf("failed to look up %s: %v", ip.String(), err)})
		return
	}
	gc.JSON(200, result)
}
