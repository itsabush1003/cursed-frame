package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/middleware"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/usecase"
)

const UploadImageFormID string = "image"

type ImageHandler struct {
	iuu        *usecase.ImageUploadUsecase
	idu        *usecase.ImageDownloadUsecase
	fileServer http.Handler
}

func (ih *ImageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	reqUser := middleware.GetUserFromCtx(r.Context())
	if reqUser == nil {
		http.Error(w, "Invalid User", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		ih.download(w, r, reqUser.GetUserID())
	case "POST":
		ih.upload(w, r, reqUser.GetUserID())
	default:
		http.Error(w, r.Method+" is not allowed", http.StatusMethodNotAllowed)
	}
}

func (ih *ImageHandler) upload(w http.ResponseWriter, r *http.Request, reqUserID uuid.UUID) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileSrc, _, err := r.FormFile(UploadImageFormID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fileSrc.Close()

	if err = ih.iuu.Execute(fileSrc, reqUserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ih *ImageHandler) download(w http.ResponseWriter, r *http.Request, reqUserID uuid.UUID) {
	imagePath, err := ih.idu.Execute(r.URL.Path, reqUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.URL.Path = imagePath
	ih.fileServer.ServeHTTP(w, r)
}

func NewImageHandler(iuu *usecase.ImageUploadUsecase, idu *usecase.ImageDownloadUsecase, dirName string) *ImageHandler {
	return &ImageHandler{
		iuu:        iuu,
		idu:        idu,
		fileServer: http.FileServer(http.Dir(dirName)),
	}
}
