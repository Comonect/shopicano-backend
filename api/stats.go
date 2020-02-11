package api

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"net/http"
	"sort"
	"time"
)

func RegisterStatsRoutes(g *echo.Group) {
	func(g *echo.Group) {
		g.Use(middlewares.MightBeStoreStaffAndStoreActive)
		g.GET("/products/", productStats)
		g.GET("/categories/", categoryStats)
		//g.GET("/collections/", collectionStats)
		//g.GET("/stores/", storeStats)
	}(g)

	func(g *echo.Group) {
		g.Use(middlewares.IsStoreStaffAndStoreActive)
		g.GET("/orders/", orderStats)
		//g.GET("/collections/", collectionStats)
		//g.GET("/stores/", storeStats)
	}(g)
}

func productStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	pu := data.NewProductRepository()

	var res interface{}
	var err error

	isPublic := !utils.IsStoreStaff(ctx)
	if isPublic {
		res, err = pu.Stats(db, 0, 25)
	} else {
		res, err = pu.StatsAsStoreStaff(db, utils.GetStoreID(ctx), 0, 25)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func categoryStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	cu := data.NewCategoryRepository()

	var res interface{}
	var err error

	isPublic := !utils.IsStoreStaff(ctx)
	if isPublic {
		res, err = cu.Stats(db, 0, 25)
	} else {
		res, err = cu.StatsAsStoreStuff(db, utils.GetStoreID(ctx), 0, 25)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func orderStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	ou := data.NewOrderRepository()

	timeFrames := map[time.Time]time.Time{}

	timeline := ctx.QueryParam("timeline")
	switch timeline {
	case "w":
		now := time.Now()
		prev := now.Add(time.Hour * -24)

		for len(timeFrames) < 7 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * -24)
		}
	case "m":
		now := time.Now()
		prev := now.Add(time.Hour * 7 * -24)

		for len(timeFrames) < 5 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * 7 * -24)
		}
	case "y":
		now := time.Now()
		prev := now.Add(time.Hour * 30 * -24)

		for len(timeFrames) < 12 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * 30 * -24)
		}
	}

	summary, err := ou.StoreSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	var timeWiseSummary []*models.Summary
	ordersStats := treemap.NewWith(func(a, b interface{}) int {
		x := a.(string)
		y := b.(string)

		xStart, _ := time.Parse(utils.DateFormat, x)
		yStart, _ := time.Parse(utils.DateFormat, y)
		if xStart.After(yStart) {
			return 1
		}
		return -1
	})

	earningsStats := treemap.NewWith(func(a, b interface{}) int {
		x := a.(string)
		y := b.(string)

		xStart, _ := time.Parse(utils.DateFormat, x)
		yStart, _ := time.Parse(utils.DateFormat, y)
		if xStart.After(yStart) {
			return 1
		}
		return -1
	})

	for k, v := range timeFrames {
		sum, err := ou.StoreSummaryByTime(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
		sum.Time = k.Format(utils.DateFormat)

		cStat, err := ou.CountByStatus(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pending := 0
		confirmed := 0
		shipping := 0
		delivered := 0
		cancelled := 0

		for _, x := range cStat {
			switch x.Key {
			case string(models.OrderPending):
				pending = x.Value
			case string(models.OrderConfirmed):
				confirmed = x.Value
			case string(models.OrderShipping):
				shipping = x.Value
			case string(models.OrderDelivered):
				delivered = x.Value
			case string(models.OrderCancelled):
				cancelled = x.Value
			}
		}

		ordersStats.Put(sum.Time, map[string]int{
			"pending":   pending,
			"confirmed": confirmed,
			"shipping":  shipping,
			"delivered": delivered,
			"cancelled": cancelled,
		})

		eStat, err := ou.EarningsByStatus(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pending = 0
		completed := 0
		failed := 0
		reverted := 0

		for _, x := range eStat {
			switch x.Key {
			case string(models.PaymentPending):
				pending = x.Value
			case string(models.PaymentCompleted):
				completed = x.Value
			case string(models.PaymentFailed):
				failed = x.Value
			case string(models.PaymentReverted):
				reverted = x.Value
			}
		}

		earningsStats.Put(sum.Time, map[string]int{
			"pending":   pending,
			"completed": completed,
			"failed":    failed,
			"reverted":  reverted,
		})

		timeWiseSummary = append(timeWiseSummary, sum)
	}

	sort.Slice(timeWiseSummary, func(i, j int) bool {
		x := timeWiseSummary[i].Time
		y := timeWiseSummary[j].Time

		xStart, _ := time.Parse(utils.DateFormat, x)
		yStart, _ := time.Parse(utils.DateFormat, y)
		return !xStart.After(yStart)
	})

	ordersIt := map[string]map[string]int{}
	ordersStats.Each(func(k interface{}, v interface{}) {
		ordersIt[k.(string)] = v.(map[string]int)
	})

	earningsIt := map[string]map[string]int{}
	earningsStats.Each(func(k interface{}, v interface{}) {
		earningsIt[k.(string)] = v.(map[string]int)
	})

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"report":           summary,
		"reports_by_time":  timeWiseSummary,
		"orders_by_time":   ordersIt,
		"earnings_by_time": earningsIt,
	}
	return resp.ServerJSON(ctx)
}
