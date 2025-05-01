package controller

import (
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	C "air-example/ui/html/component"
	L "air-example/ui/html/layout"
)

func BaseHandler(c echo.Context) error {
	component := L.Base("Test")
	ctx := templ.WithChildren(context.Background(), C.Hello("World"))

	return component.Render(ctx, c.Response().Writer)
}
