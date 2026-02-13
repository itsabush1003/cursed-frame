package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

const HeaderKey string = "Authorization"

type UserContextKey struct{}

func GetUserFromCtx(ctx context.Context) *model.User {
	return ctx.Value(UserContextKey{}).(*model.User)
}

type IUserRepository interface {
	FetchByToken(string) (*model.User, error)
}

type AuthorizeMiddleware struct {
	ur IUserRepository
}

func (am *AuthorizeMiddleware) authByRequestHeader(header string) (*model.User, error) {
	if header == "" {
		return nil, errors.New("Authorization header is required")
	}
	token := strings.TrimPrefix(header, "Bearer ")

	user, err := am.ur.FetchByToken(token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (am *AuthorizeMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(HeaderKey)

		user, err := am.authByRequestHeader(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey{}, user)))
	})
}

func (am *AuthorizeMiddleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		authHeader := request.Header().Get(HeaderKey)

		user, err := am.authByRequestHeader(authHeader)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		return next(context.WithValue(ctx, UserContextKey{}, user), request)
	}
}

func (am *AuthorizeMiddleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		authHeader := conn.RequestHeader().Get(HeaderKey)
		user, err := am.authByRequestHeader(authHeader)
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, err)
		}

		return next(context.WithValue(ctx, UserContextKey{}, user), conn)
	}
}

func (am *AuthorizeMiddleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		// ClientStreamingは使ってないので素通し
		conn := next(ctx, spec)
		return conn
	}
}

func NewAuthorizeMiddleware(ur IUserRepository) *AuthorizeMiddleware {
	return &AuthorizeMiddleware{
		ur: ur,
	}
}
