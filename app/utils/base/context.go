package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"unicode"

	ci18n "mcop/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	msg "mcop/config/i18n"
)

// Regexp definitions
var (
	keyMatchRegex    = regexp.MustCompile(`\"(\w+)\":`)
	wordBarrierRegex = regexp.MustCompile(`([a-z_0-9])([A-Z])`)
)

type conventionalMarshallerFromPascal struct {
	Value any
}

func convertToCamelCase(marshalled []byte) []byte {
	return keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			// log.Info("Original:", string(match))
			// Empty keys are valid JSON, only lowercase if we do not have an empty key.
			if len(match) > 2 {
				// Convert to camel case
				converted := bytes.ToLower(wordBarrierRegex.ReplaceAll(
					match,
					[]byte(`${1}_${2}`),
				))

				// Remove underscores and capitalize the following letter
				var result []byte
				underscore := false
				for i := 1; i < len(converted)-1; i++ {
					if converted[i] == '_' {
						underscore = true
					} else {
						if underscore {
							result = append(result, byte(unicode.ToUpper(rune(converted[i]))))
							underscore = false
						} else {
							result = append(result, converted[i])
						}
					}
				}
				result = append([]byte{converted[0]}, result...)
				result = append(result, converted[len(converted)-1])
				// log.Info("Converted:", string(result))
				return result
			}
			return match
		},
	)
}

func (c conventionalMarshallerFromPascal) MarshalJSON() ([]byte, error) {
	marshalled, err := json.Marshal(c.Value)
	if err != nil {
		return nil, err
	}
	naming, ok := os.LookupEnv("HTTP_JSON_NAMING")
	if !ok {
		naming = "snake_case"
	}

	val := reflect.TypeOf(c.Value)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {

		field, ok := val.FieldByName("json")
		if ok {
			if field.Tag.Get("naming") != "" {
				naming = field.Tag.Get("naming")
			}
		}
	}

	var converted []byte
	switch naming {
	case "snake_case":

		// https://gist.github.com/Rican7/39a3dc10c1499384ca91
		converted = keyMatchRegex.ReplaceAllFunc(
			marshalled,
			func(match []byte) []byte {
				return bytes.ToLower(wordBarrierRegex.ReplaceAll(
					match,
					[]byte(`${1}_${2}`),
				))
			},
		)
	case "camel_case":
		converted = convertToCamelCase(marshalled)
	case "pascal_case":
		return marshalled, nil
	default:
		return nil, err
	}

	return converted, nil
}

// JSON sends a JSON response with the given status code, message ID, data, and pagination information.
// It supports localization of messages and custom parameter substitution.
func JSON(ctx *gin.Context, code int, msgID string, data any, paginate *ResponsePaginate, params ...map[string]string) error {
	localizer := i18n.NewLocalizer(ci18n.Bundle, ctx.GetHeader("Accept-Language"))

	var param any
	if len(params) > 0 {
		param = params[0]
	}
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: msgID, TemplateData: param})

	if err != nil || msg == "" {
		ctx.JSON(code, Response[any]{
			ResponseStatus: &ResponseStatus{
				Message: msgID,
				Code:    fmt.Sprintf("%d", code),
			},
			Data:     data,
			Paginate: paginate,
		})
		return nil
	}

	ctx.JSON(code, conventionalMarshallerFromPascal{Response[any]{
		ResponseStatus: &ResponseStatus{
			Message: msg,
			Code:    strconv.Itoa(code),
		},
		Data:     data,
		Paginate: paginate,
	}})
	return nil
}

// RawJSON sends a JSON response with the given status code and data using conventional marshalling
func RawJSON(ctx *gin.Context, code int, data any) error {
	ctx.JSON(code, conventionalMarshallerFromPascal{data})
	return nil
}

// Success 200 success
func Success(ctx *gin.Context, data any, message ...string) error {
	msg := msg.Success
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return JSON(ctx, http.StatusOK, msg, data, nil)
}

// Paginate 200 success
func Paginate(ctx *gin.Context, data any, page *ResponsePaginate) error {
	return JSON(ctx, http.StatusOK, msg.Success, data, page)
}

// Created 201 created
func Created(ctx *gin.Context, message string) error {
	return JSON(ctx, http.StatusCreated, message, nil, nil)
}

// BadRequest 400 other and external error
func BadRequest(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusBadRequest, message, data, nil, params...)
}

// Unauthorized 401 un authentication
func Unauthorized(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusUnauthorized, message, data, nil, params...)
}

// Forbidden 403 No permission
func Forbidden(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusForbidden, message, data, nil, params...)
}

// ValidateFailed 412 Validate error
func ValidateFailed(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusPreconditionFailed, message, data, nil, params...)
}

// InternalServerError 500 internal error
func InternalServerError(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusInternalServerError, message, data, nil, params...)
}

// NotImplemented 501 not implemented
func NotImplemented(ctx *gin.Context, message string, data any, params ...map[string]string) error {
	return JSON(ctx, http.StatusNotImplemented, message, data, nil, params...)
}
