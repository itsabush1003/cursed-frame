package usecase

import (
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

type ImageUploadUsecase struct {
	imgDirName string
	uir IUserImageRepository
}

func (iuu *ImageUploadUsecase) Execute(fileSrc io.Reader, uid uuid.UUID) error {
	filaName, err := util.CreateRandStr(8)
	if err != nil {
		return err
	}
	fileDest, err := os.Create(fmt.Sprintf("%s/%s.jpg", iuu.imgDirName, filaName))
	if err != nil {
		return err
	}
	defer fileDest.Close()

	if _, err := io.Copy(fileDest, fileSrc); err != nil {
		return err
	}

	iuu.uir.Save(uid, filaName)

	return nil
}

func NewImageUploadUsecase(imgDirName string, uir IUserImageRepository) *ImageUploadUsecase {
	return &ImageUploadUsecase{
		imgDirName: imgDirName,
		uir: uir,
	}
}
