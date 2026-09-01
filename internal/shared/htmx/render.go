package htmx

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

// Render writes component as the response body — every feature handler's
// terminal step, shared here so it isn't redefined per package.
func Render(c fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
