package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterProductRoutes(g *echo.Group) {
	func(g *echo.Group) {
		g.Use(middlewares.MightBeStoreStaffWithStoreActivation)
		g.GET("/", listProducts)
		g.GET("/:product_id/", getProduct)
	}(g)

	func(g *echo.Group) {
		// Private endpoints only
		g.Use(middlewares.IsStoreStaffWithStoreActivation)
		g.POST("/", createProduct)
		g.PATCH("/:product_id/", updateProduct)
		g.DELETE("/:product_id/", deleteProduct)
		g.PUT("/:product_id/attributes/", addProductAttribute)
		g.DELETE("/:product_id/attributes/:attribute_key/", deleteProductAttribute)
	}(g)
}

func createProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	req, err := validators.ValidateCreateProduct(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	req.StoreID = storeID
	if req.CategoryID != nil && *req.CategoryID == "" {
		req.CategoryID = nil
	}

	images := ""
	for _, i := range req.AdditionalImages {
		if strings.TrimSpace(i) == "" {
			continue
		}
		if images != "" {
			images += ","
		}
		images += strings.TrimSpace(i)
	}

	p := models.Product{
		ID:                  utils.NewUUID(),
		StoreID:             req.StoreID,
		Price:               req.Price,
		Stock:               req.Stock,
		Name:                req.Name,
		IsShippable:         req.IsShippable,
		CategoryID:          req.CategoryID,
		IsPublished:         req.IsPublished,
		IsDigital:           req.IsDigital,
		AdditionalImages:    images,
		SKU:                 req.SKU,
		Unit:                req.Unit,
		DigitalDownloadLink: req.DigitalDownloadLink,
		Image:               req.Image,
		Description:         req.Description,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	db := app.DB()

	pu := data.NewProductRepository()
	err = pu.Create(db, &p)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.ProductAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Title = "Product created"
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func updateProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	req, err := validators.ValidateUpdateProduct(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	req.StoreID = storeID
	if req.CategoryID != nil && *req.CategoryID == "" {
		req.CategoryID = nil
	}

	db := app.DB()

	pu := data.NewProductRepository()

	images := ""
	for _, i := range req.AdditionalImages {
		if strings.TrimSpace(i) == "" {
			continue
		}
		if images != "" {
			images += ","
		}
		images += strings.TrimSpace(i)
	}

	p, err := pu.Get(db, productID)
	if err != nil {
		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	p.ID = productID
	p.StoreID = req.StoreID
	p.Price = req.Price
	p.Stock = req.Stock
	p.Name = req.Name
	p.IsShippable = req.IsShippable
	p.CategoryID = req.CategoryID
	p.IsPublished = req.IsPublished
	p.IsDigital = req.IsDigital
	p.AdditionalImages = images
	p.SKU = req.SKU
	p.Unit = req.Unit
	p.DigitalDownloadLink = req.DigitalDownloadLink
	p.Image = req.Image
	p.Description = req.Description
	p.UpdatedAt = time.Now().UTC()

	err = pu.Update(db, productID, p)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.ProductAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Product updated"
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func deleteProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()
	if err := pu.Delete(db, storeID, productID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func getProduct(ctx echo.Context) error {
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()

	var p interface{}
	var err error

	if utils.IsStoreStaff(ctx) {
		p, err = pu.GetAsStoreStuff(db, ctx.Get(utils.StoreID).(string), productID)
	} else {
		p, err = pu.GetDetails(db, productID)
	}

	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func listProducts(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	var r interface{}

	if query == "" {
		r, err = fetchProducts(ctx, page, limit, !utils.IsStoreStaff(ctx))
	} else {
		r, err = searchProducts(ctx, query, page, limit, !utils.IsStoreStaff(ctx))
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = r
	return resp.ServerJSON(ctx)
}

func fetchProducts(ctx echo.Context, page int64, limit int64, isPublic bool) (interface{}, error) {
	from := (page - 1) * limit
	pu := data.NewProductRepository()

	db := app.DB()

	if isPublic {
		return pu.List(db, int(from), int(limit))
	}
	return pu.ListAsStoreStuff(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func searchProducts(ctx echo.Context, query string, page int64, limit int64, isPublic bool) (interface{}, error) {
	from := (page - 1) * limit
	pu := data.NewProductRepository()

	db := app.DB()

	if isPublic {
		return pu.Search(db, query, int(from), int(limit))
	}
	return pu.SearchAsStoreStuff(db, query, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func addProductAttribute(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	resp := core.Response{}

	req, err := validators.ValidateAddProductAttribute(ctx)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductAttributeCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	pu := data.NewProductRepository()

	p, err := pu.GetAsStoreStuff(db, storeID, productID)
	if err != nil {
		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	v := models.ProductAttribute{
		ProductID: p.ID,
		Key:       req.Key,
		Value:     req.Value,
	}

	err = pu.AddAttribute(db, &v)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.ProductAttributeAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Product attribute added"
	return resp.ServerJSON(ctx)
}

func deleteProductAttribute(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")
	attributeKey := ctx.Param("attribute_key")

	resp := core.Response{}

	db := app.DB()
	pu := data.NewProductRepository()

	p, err := pu.GetAsStoreStuff(db, storeID, productID)
	if err != nil {
		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	err = pu.RemoveAttribute(db, p.ID, attributeKey)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}
