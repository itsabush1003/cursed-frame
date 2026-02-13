package infra

import (
	"fmt"
	"net/http"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/admin/v1/adminv1connect"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/entry/v1/entryv1connect"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/lobby/v1/lobbyv1connect"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/quest/v1/questv1connect"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/go-pkgz/routegroup"

	filecontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/file"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/middleware"
	restcontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/rest"
	rpccontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/rpc"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

const RandomPathLength int = 6

func redirectHandlerFunc(path string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusSeeOther)
	})
}

type Router struct {
	router    *routegroup.Bundle
	AdminPath string
	GuestPath string
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.router.ServeHTTP(w, r)
}

func NewRouter(
	fileHandler *filecontroller.StaticFileHandler,
	imageHndler *restcontroller.ImageHandler,
	entryServiceHandler *rpccontroller.EntryServiceHandler,
	lobbyServiceHandler *rpccontroller.LobbyServiceHandler,
	questServiceHandler *rpccontroller.QuestServiceHandler,
	adminServiceHandler *rpccontroller.AdminServiceHandler,
	authorizeMiddleware *middleware.AuthorizeMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	corsMiddleware *middleware.CorsMiddleware,
) *Router {
	// adminとguest２つ分のランダム文字列をまとめて生成して後で分割
	randStr, err := util.CreateRandStr(RandomPathLength * 2)
	if err != nil {
		panic(err)
	}
	runes := []rune(randStr)
	adminPath := fmt.Sprintf("/%s/admin", string(runes[0:len(runes)/2]))
	guestPath := fmt.Sprintf("/%s/guest", string(runes[len(runes)/2:]))

	router := routegroup.New(http.NewServeMux())
	adminGroup := router.Mount(adminPath)
	guestGroup := router.Mount(guestPath)
	adminGroup.Use(corsMiddleware.Handle)
	guestGroup.Use(rateLimitMiddleware.Handle, corsMiddleware.Handle)
	adminGroup.HandleFunc("/", redirectHandlerFunc("static"))
	guestGroup.HandleFunc("/", redirectHandlerFunc("static"))
	adminGroup.Handle("/static/", http.StripPrefix(adminPath+"/static", http.HandlerFunc(fileHandler.Handle)))
	guestGroup.Handle("/static/", http.StripPrefix(guestPath+"/static", http.HandlerFunc(fileHandler.Handle)))
	adminRestGroup := adminGroup.Mount("/rest")
	guestRestGroup := guestGroup.Mount("/rest")
	guestRestGroup.Use(authorizeMiddleware.Handle)
	adminRestGroup.Handle("GET /images", http.StripPrefix(adminPath+"/rest/images", http.HandlerFunc(imageHndler.Handle)))
	guestRestGroup.Handle("GET /images", http.StripPrefix(guestPath+"/rest/images", http.HandlerFunc(imageHndler.Handle)))
	// imageをuploadする必要があるのはゲストだけ
	guestRestGroup.Handle("POST /images", http.StripPrefix(guestPath+"/rest/images", http.HandlerFunc(imageHndler.Handle)))

	guestRpcGroup := guestGroup.Mount("/rpc")
	entryPath, entryHandler := entryv1connect.NewEntryServiceHandler(
		entryServiceHandler,
		// Validation via Protovalidate is almost always recommended
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	guestRpcGroup.Handle(entryPath, http.StripPrefix(guestPath+"/rpc", entryHandler))
	lobbyPath, lobbyHandler := lobbyv1connect.NewLobbyServiceHandler(
		lobbyServiceHandler,
		// Validation via Protovalidate is almost always recommended
		connect.WithInterceptors(validate.NewInterceptor(), authorizeMiddleware),
	)
	guestRpcGroup.Handle(lobbyPath, http.StripPrefix(guestPath+"/rpc", lobbyHandler))
	questPath, questHandler := questv1connect.NewQuestServiceHandler(
		questServiceHandler,
		// Validation via Protovalidate is almost always recommended
		connect.WithInterceptors(validate.NewInterceptor(), authorizeMiddleware),
	)
	guestRpcGroup.Handle(questPath, http.StripPrefix(guestPath+"/rpc", questHandler))

	adminRpcGroup := adminGroup.Mount("/rpc")
	adminsPath, adminHandler := adminv1connect.NewAdminServiceHandler(
		adminServiceHandler,
		// Validation via Protovalidate is almost always recommended
		connect.WithInterceptors(validate.NewInterceptor(), authorizeMiddleware),
	)
	adminRpcGroup.Handle(adminsPath, http.StripPrefix(adminPath+"/rpc", adminHandler))

	return &Router{
		router:    router,
		AdminPath: adminPath,
		GuestPath: guestPath,
	}
}
