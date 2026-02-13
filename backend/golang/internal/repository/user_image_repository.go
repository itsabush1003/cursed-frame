package repository

import "github.com/google/uuid"

type UserImageRepository struct {
	db IDatabase
}

func (uir *UserImageRepository) Save(uid uuid.UUID, imageID string) error {
	resultCh := make(chan error, 1)
	uir.db.Command("UserAttribute", WriteRequest{
		Table:   "UserImage",
		Method:  Insert,
		Targets: []string{"user_id", "image_id"},
		Params: map[string]string{
			"user_id":  uid.String(),
			"image_id": imageID,
		},
		Conds:    "",
		ResultCh: resultCh,
	})
	if err := <-resultCh; err != nil {
		return err
	}
	return nil
}

func (uir *UserImageRepository) FetchByUserID(uid uuid.UUID) (string, error) {
	var imageID string
	if err := uir.db.QueryRow("UserAttribute", "SELECT image_id FROM UserImage WHERE user_id = ?", uid.String()).Scan(&imageID); err != nil {
		return "", err
	}
	return imageID, nil
}

func NewUserImageRepository(db IDatabase) *UserImageRepository {
	return &UserImageRepository{
		db: db,
	}
}
