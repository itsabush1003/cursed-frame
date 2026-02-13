package middleware

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

type AdminCheckMiddleware struct {
	ur IUserRepository
}

func (acm *AdminCheckMiddleware) checkAdmin(ctx context.Context) error {
	user := GetUserFromCtx(ctx)
	if user == nil {
		return errors.New("Unauthorized Access")
	}
	return nil
}

func (acm *AdminCheckMiddleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if err := acm.checkAdmin(ctx); err != nil {
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

func NewAdminCheckMiddleware(ur IUserRepository) *AuthorizeMiddleware {
	return &AuthorizeMiddleware{
		ur: ur,
	}
}
