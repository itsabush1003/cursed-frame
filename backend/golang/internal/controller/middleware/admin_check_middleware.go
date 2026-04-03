package middleware

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

type AdminCheckMiddleware struct {
	adminUserId string
}

func (acm *AdminCheckMiddleware) checkAdmin(ctx context.Context) error {
	user := GetUserFromCtx(ctx)
	if user == nil {
		return errors.New("You are Unauthorized")
	}
	if acm.adminUserId == "" {
		acm.adminUserId = user.GetUserID().String()
		fmt.Println("Administrator has registered: ", acm.adminUserId[:4]+"xxxx-x...")
	} else if acm.adminUserId != user.GetUserID().String() {
		return errors.New("You are not administrator")
	}
	return nil
}

func (acm *AdminCheckMiddleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if err := acm.checkAdmin(ctx); err != nil {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return next(ctx, request)
	}
}

func (acm *AdminCheckMiddleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		if err := acm.checkAdmin(ctx); err != nil {
			return connect.NewError(connect.CodePermissionDenied, err)
		}

		return next(ctx, conn)
	}
}

func (acm *AdminCheckMiddleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	}
}

func NewAdminCheckMiddleware() *AdminCheckMiddleware {
	return &AdminCheckMiddleware{
		adminUserId: "",
	}
}
