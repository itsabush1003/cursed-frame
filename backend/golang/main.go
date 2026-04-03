package main

import (
	"crypto/tls"
	"embed"
	"encoding/base64"
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

const SecretLength int = 16
const TempDirName string = "user_images"
const EnvPrefix string = "PCF_"

var (
	userNum     int
	teamNum     int
	certFile    string
	keyFile     string
	domain      string
	useAutoCert bool
)

//go:embed dist/*
var staticFiles embed.FS

//go:embed migration/*
var dbSourceFiles embed.FS

func init() {
	flag.IntVar(&userNum, "N", 6, "総参加者数")
	flag.IntVar(&teamNum, "T", 2, "参加者を振り分けるチーム数")
	flag.StringVar(&certFile, "cert", os.Getenv(EnvPrefix+"SSL_CERT_FILE"), "TLS用証明書ファイル")
	flag.StringVar(&keyFile, "key", os.Getenv(EnvPrefix+"SSL_KEY_FILE"), "TLS用鍵ファイル")
	flag.StringVar(&domain, "domain", os.Getenv(EnvPrefix+"DOMAIN"), "ドメイン")
	flag.BoolVar(&useAutoCert, "autocert", false, "証明書の自動生成を有効にするか")
}

func main() {
	flag.Parse()

	secret, err := util.CreateRandStr(SecretLength)
	if err != nil {
		panic(err)
	}
	byteSecret, err := base64.RawURLEncoding.DecodeString(secret)
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

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if certFile != "" && keyFile != "" {
		certs, err := infra.CreateCertificateFromFiles(certFile, keyFile)
		if err != nil {
			panic(err)
		}
		tlsConfig.Certificates = certs
	} else if useAutoCert {
		getCertificate, cleanAutoCert, err := infra.CreateCertificateWithAutoCert(domain)
		if err != nil {
			panic(err)
		}
		defer cleanAutoCert()
		tlsConfig.GetCertificate = getCertificate
	}

	// "dist"ディレクトリをルートとして扱う
	dist, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}
	dbSource, err := fs.Sub(dbSourceFiles, "migration")
	if err != nil {
		panic(err)
	}

	gameManager := core.NewGameManager(userNum, teamNum)
	database, err := infra.NewSQLiteDB(dbDirname, dbSource)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	// 初期化 TODO: DIにする
	fileHandler := filecontroller.NewStaticFileHandler(http.FS(dist))
	c := cache.New(10*time.Minute, 30*time.Minute)
	userRepository := repository.NewUserRepository(c, database)
	adminCheckMiddleware := middleware.NewAdminCheckMiddleware()
	authorizeMiddleware := middleware.NewAuthorizeMiddleware(userRepository)
	corsMiddleware := middleware.NewCorsMiddleware()
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(userNum)
	userImageRepository := repository.NewUserImageRepository(database)
	imageUploadUsecase := usecase.NewImageUploadUsecase(imageDirname, userImageRepository)
	imageDownloadUsecase := usecase.NewImageDownloadUsecase(userImageRepository)
	imageHandler := restcontroller.NewImageHandler(imageUploadUsecase, imageDownloadUsecase, imageDirname)
	entryUsecase := usecase.NewEntryUsecase(userRepository, byteSecret)
	reconnectUsecase := usecase.NewReconnectUsecase(byteSecret, userRepository)
	entryServiceHandler := rpccontroller.NewEntryServiceHandler(entryUsecase, reconnectUsecase)
	profileQuestionRepository := repository.NewProfileQuestionRepository(database)
	userProfileRepository := repository.NewUserProfileRepository(database)
	joinLobbyUsecase := usecase.NewJoinLobbyUsecase(gameManager)
	registProfileUsecase := usecase.NewRegistProfileUsecase(profileQuestionRepository, userProfileRepository)
	setReadyUsecase := usecase.NewSetReadyUsecase(userRepository)
	getTeamInfoUsecase := usecase.NewGetTeamInfoUsecase(userRepository)
	lobbyServiceHandler := rpccontroller.NewLobbyServiceHandler(joinLobbyUsecase, registProfileUsecase, setReadyUsecase, getTeamInfoUsecase)
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
	readyQuizUsecase := usecase.NewReadyQuizUsecase(gameManager)
	checkAnswersUsecase := usecase.NewCheckAnswersUsecase(gameManager)
	nextQuizUsecase := usecase.NewNextQuizUsecase(gameManager)
	endQuestUsecase := usecase.NewEndQuestUsecase(gameManager, userRepository, infra.ResultStateMapper)
	adminServiceHandler := rpccontroller.NewAdminServiceHandler(openEntryUsecase, closeEntryUsecase, rejectUserUsecase, changeTeamUsecase, adminStartQuestUsecase, readyQuizUsecase, checkAnswersUsecase, nextQuizUsecase, endQuestUsecase, userNum)
	router := infra.NewRouter(fileHandler, imageHandler, entryServiceHandler, lobbyServiceHandler, questServiceHandler, adminServiceHandler, adminCheckMiddleware, authorizeMiddleware, rateLimitMiddleware, corsMiddleware)

	server := infra.NewServer(":8888", tlsConfig, router)
	println(fmt.Sprintf("Server started at :8888%s\n\tguest: %s", router.AdminPath, router.GuestPath))
	if err = server.ListenAndServe(); err != nil {
		panic(err)
	}
}
