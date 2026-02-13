package usecase

import (
	"strings"

	"github.com/google/uuid"
)

const ImageFileExtension string = ".jpg"

type ImageDownloadUsecase struct {
	uir IUserImageRepository
}

func (idu *ImageDownloadUsecase) Execute(imagePath string, uid uuid.UUID) (string, error) {
	if strings.HasSuffix(imagePath, "/") {
		imageID, err := idu.uir.FetchByUserID(uid)
		if err != nil {
			return "", err
		}
		imagePath = imagePath + imageID + ImageFileExtension
		return imagePath, nil
	} else if !strings.HasSuffix(imagePath, ImageFileExtension) {
		return imagePath + ImageFileExtension, nil
	}

	return imagePath, nil
}

func NewImageDownloadUsecase(uir IUserImageRepository) *ImageDownloadUsecase {
	return &ImageDownloadUsecase{
		uir: uir,
	}
}
