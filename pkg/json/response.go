package json_response

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SonicJSON struct {
	Data interface{}
}

func (r SonicJSON) Render(w http.ResponseWriter) error {
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func (r SonicJSON) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// var jsonAPI = sonic.Config{
// 	UseNumber:            true,
// 	EscapeHTML:           false,
// 	SortMapKeys:          false,
// 	CompactMarshaler:     true,
// 	NoQuoteTextMarshaler: true,
// 	NoNullSliceOrMap:     true,
// }.Froze()

var (
	successResponse       = mustMarshal(Response{Code: 200, Message: "Success"})
	createdResponse       = mustMarshal(Response{Code: 201, Message: "Created"})
	notFoundResponse      = mustMarshal(Response{Code: 404, Message: "Not Found"})
	unauthorizedResponse  = mustMarshal(Response{Code: 401, Message: "Unauthorized"})
	badRequestResponse    = mustMarshal(Response{Code: 400, Message: "Bad Request"})
	forbiddenResponse     = mustMarshal(Response{Code: 403, Message: "Forbidden"})
	internalErrorResponse = mustMarshal(Response{Code: 500, Message: "Internal Server Error"})
)

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func ResponseJSON(c *fiber.Ctx, httpCode int, message string, data interface{}) {
	if data == nil {
		switch httpCode {
		case 200:
			if message == "Success" {
				c.Status(httpCode).Send(successResponse)
				return
			}
		case 201:
			if message == "Created" {
				c.Status(httpCode).Send(createdResponse)
				return
			}
		case 400:
			if message == "Bad Request" {
				c.Status(httpCode).Send(badRequestResponse)
				return
			}
		case 404:
			if message == "Not Found" {
				c.Status(httpCode).Send(notFoundResponse)
				return
			}
		case 401:
			if message == "Unauthorized" {
				c.Status(httpCode).Send(unauthorizedResponse)
				return
			}
		case 403:
			if message == "Forbidden" {
				c.Status(httpCode).Send(forbiddenResponse)
				return
			}
		case 500:
			if message == "Internal Server Error" {
				c.Status(httpCode).Send(internalErrorResponse)
				return
			}
		}
	}

	c.Status(httpCode).JSON(Response{
		Code:    httpCode,
		Message: message,
		Data:    data,
	})
}

func ResponseOK(c *fiber.Ctx, data interface{}) {
	ResponseJSON(c, 200, "Success", data)
}

func ResponseNotFound(c *fiber.Ctx) {
	ResponseJSON(c, 404, "Not Found", nil)
}

func ResponseUnauthorized(c *fiber.Ctx) {
	ResponseJSON(c, 401, "Unauthorized", nil)
}

func ResponseBadRequest(c *fiber.Ctx, message string) {
	if message == "" {
		message = "Bad Request"
	}
	ResponseJSON(c, 400, message, nil)
}

func ResponseForbidden(c *fiber.Ctx) {
	ResponseJSON(c, 403, "Forbidden", nil)
}

func ResponseCreated(c *fiber.Ctx, data interface{}) {
	ResponseJSON(c, 201, "Created", data)
}

func ResponseInternalError(c *fiber.Ctx, err error) {
	ResponseJSON(c, 500, "Internal Server Error", err.Error())
}
