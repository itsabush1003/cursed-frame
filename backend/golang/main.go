package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"

	filecontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/file"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/middleware"
	restcontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/rest"
	rpccontroller "github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/rpc"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/infra"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/repository"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/usecase"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

const SecretLength int = 8
const TempDirName string = "user_images"

var userNum int
var teamNum int

//go:embed dist/*
var staticFiles embed.FS

func init() {
	flag.IntVar(&userNum, "N", 6, "総参加者数")
	flag.IntVar(&teamNum, "T", 2, "参加者を振り分けるチーム数")
}

func main() {
	flag.Parse()

	secret, err := util.CreateRandStr(SecretLength)
	if err != nil {
		panic(err)
	}

	imageDirname, err := os.MkdirTemp("", TempDirName)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(imageDirname)
	println(imageDirname)

	dbDirname, err := os.MkdirTemp("", "db")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dbDirname)
	println(dbDirname)

	// "dist"ディレクトリをルートとして扱う
	dist, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}

	gameManager := core.NewGameManager(userNum, teamNum)
	database, err := infra.NewSQLiteDB(dbDirname)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	// 初期化 TODO: DIにする
	fileHandler := filecontroller.NewStaticFileHandler(http.FS(dist))
	c := cache.New(10*time.Minute, 30*time.Minute)
	userRepository := repository.NewUserRepository(c)
	authorizeMiddleware := middleware.NewAuthorizeMiddleware(userRepository)
	corsMiddleware := middleware.NewCorsMiddleware()
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(userNum)
	userImageRepository := repository.NewUserImageRepository(database)
	imageUploadUsecase := usecase.NewImageUploadUsecase(imageDirname, userImageRepository)
	imageDownloadUsecase := usecase.NewImageDownloadUsecase(userImageRepository)
	imageHandler := restcontroller.NewImageHandler(imageUploadUsecase, imageDownloadUsecase, imageDirname)
	entryUsecase := usecase.NewEntryUsecase(userRepository)
	reconnectUsecase := usecase.NewReconnectUsecase(secret, userRepository)
	entryServiceHandler := rpccontroller.NewEntryServiceHandler(entryUsecase, reconnectUsecase, secret)
	profileQuestionRepository := repository.NewProfileQuestionRepository(database)
	userProfileRepository := repository.NewUserProfileRepository(database)
	joinLobbyUsecase := usecase.NewJoinLobbyUsecase(gameManager)
	registProfileUsecase := usecase.NewRegistProfileUsecase(profileQuestionRepository, userProfileRepository)
	setReadyUsecase := usecase.NewSetReadyUsecase(userRepository)
	getTeamIDUsecase := usecase.NewGetTeamIDUsecase()
	lobbyServiceHandler := rpccontroller.NewLobbyServiceHandler(joinLobbyUsecase, registProfileUsecase, setReadyUsecase, getTeamIDUsecase)
	guestStartQuestUsecase := usecase.NewGuestStartQuestUsecase(gameManager)
	answerUsecase := usecase.NewAnswerUsecase(gameManager)
	takeHintUsecase := usecase.NewTakeHintUsecase(gameManager)
	getResultUsecase := usecase.NewGetResultUsecase(gameManager, infra.ResultStateMapper)
	questServiceHandler := rpccontroller.NewQuestServiceHandler(guestStartQuestUsecase, answerUsecase, takeHintUsecase, getResultUsecase)
	openEntryUsecase := usecase.NewOpenEntryUsecase(gameManager, userRepository)
	closeEntryUsecase := usecase.NewCloseEntryUsecase(gameManager, userRepository, teamNum)
	rejectUserUsecase := usecase.NewRejectUserUsecase(gameManager, userRepository)
	changeTeamUsecase := usecase.NewChangeTeamUsecase(userRepository)
	adminStartQuestUsecase := usecase.NewAdminStartQuestUsecase(gameManager, userRepository, userImageRepository, userProfileRepository, profileQuestionRepository)
	checkAnswersUsecase := usecase.NewCheckAnswersUsecase(gameManager)
	nextQuizUsecase := usecase.NewNextQuizUsecase(gameManager)
	endQuestUsecase := usecase.NewEndQuestUsecase(gameManager, userRepository, infra.ResultStateMapper)
	adminServiceHandler := rpccontroller.NewAdminServiceHandler(openEntryUsecase, closeEntryUsecase, rejectUserUsecase, changeTeamUsecase, adminStartQuestUsecase, checkAnswersUsecase, nextQuizUsecase, endQuestUsecase)
	router := infra.NewRouter(fileHandler, imageHandler, entryServiceHandler, lobbyServiceHandler, questServiceHandler, adminServiceHandler, authorizeMiddleware, rateLimitMiddleware, corsMiddleware)

	println(fmt.Sprintf("Server started at :8888%s\n\tguest: %s", router.AdminPath, router.GuestPath))
	http.ListenAndServe(":8888", router)
}
