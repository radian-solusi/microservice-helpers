package web

import "github.com/gin-gonic/gin"

const (
	NoCode              = 0
	Success             = 200
	SuccessCreate       = 201
	SuccessNoContent    = 204
	BadRequest          = 400
	Unauthorized        = 401
	Forbidden           = 403
	NotFound            = 404
	MethodNotAllowed    = 405
	Conflict            = 409
	ValidationError     = 422
	TooManyRequest      = 429
	InternalServerError = 500
	ServiceBroken       = 502
	ServiceUnavailable  = 503
	GatewayTimeout      = 504
	Expired             = 419
	WrongCredential     = 1001
)

func ErrorMessage(code int) string {
	switch code {
	case BadRequest:
		return "Bad Request"
	case Unauthorized:
		return "Unauthorized"
	case Forbidden:
		return "Forbidden"
	case NotFound:
		return "Not Found"
	case MethodNotAllowed:
		return "Method Not Allowed"
	case InternalServerError:
		return "Internal Server Error"
	case ServiceUnavailable:
		return "Service Unavailable"
	case GatewayTimeout:
		return "Gateway Timeout"
	case ServiceBroken:
		return "Service Not Completed"
	case WrongCredential:
		return "Username or Password is incorrect"
	case TooManyRequest:
		return "Too Many Request"
	}
	return ""
}

func ErrorResponse(ctx *gin.Context, code int, message string) {
	if code == WrongCredential {
		code = Unauthorized
	}
	SendResponse(ctx, ResponseDefault{Status: false, Code: code, Data: nil, Message: message})
	ctx.Abort()
}

type ErrorCodeMapper func(error) int

func HandleErrorResponse(ctx *gin.Context, err error, mapper ErrorCodeMapper) {
	code := InternalServerError
	if mapper != nil {
		if mapped := mapper(err); mapped != 0 {
			code = mapped
		}
	}
	ErrorResponse(ctx, code, err.Error())
}
