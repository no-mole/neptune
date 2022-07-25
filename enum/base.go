package enum

import (
	"net/http"
)

var (
	IllegalParam ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusBadRequest,
		Code:     400,
		Msg:      "illegal parameter",
	}

	Unauthorized ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusUnauthorized,
		Code:     401,
		Msg:      "unauthorized",
	}

	Success ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusOK,
		Code:     200,
		Msg:      "success",
	}

	Forbidden ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusForbidden,
		Code:     403,
		Msg:      "forbidden",
	}

	ErrorExecuteFunc ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusBadRequest,
		Code:     500,
		Msg:      "execute func error",
	}

	ErrorGrpcConnect ErrorNum = ErrorNumEntry{
		HttpCode: http.StatusInternalServerError,
		Code:     500,
		Msg:      "grpc connect error",
	}
)
