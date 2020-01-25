package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqOrderItem struct {
	ID       string `json:"id" valid:"required"`
	Quantity int    `json:"quantity" valid:"range(1|10000000)"`
}

type ReqOrderCreate struct {
	Items             []ReqOrderItem `json:"items" valid:"required"`
	ShippingAddressID *string        `json:"shipping_address_id"`
	BillingAddressID  string         `json:"billing_address_id" valid:"required"`
	PaymentMethodID   string         `json:"payment_method_id" valid:"required"`
	ShippingMethodID  *string        `json:"shipping_method_id"`
	UserID            string         `json:"user_id"`
	CouponCode        *string        `json:"coupon_code"`
}

func ValidateCreateOrder(ctx echo.Context) (*ReqOrderCreate, error) {
	pld := ReqOrderCreate{}
	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &pld, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}
