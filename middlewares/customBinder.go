package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
)

// Custom Binder
type CustomBinder struct{}

// Bind method enforces strict JSON key validation
func (cb *CustomBinder) Bind(i interface{}, c echo.Context) error {
	// Get request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	// Decode with `DisallowUnknownFields()`
	decoder := json.NewDecoder(io.NopCloser(io.MultiReader(io.NewSectionReader(bytes.NewReader(body), 0, int64(len(body))))))
	decoder.DisallowUnknownFields() // 🚀 Strict JSON validation

	if err := decoder.Decode(i); err != nil {
		return fmt.Errorf("invalid JSON: check field names")
	}

	return nil
}
