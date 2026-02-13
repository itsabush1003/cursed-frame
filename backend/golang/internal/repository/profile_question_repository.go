package repository

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"

type DBQuestionRow struct {
	QuestionID   int    `db:"question_id"`
	QuestionText string `db:"question_text"`
	QuizText     string `db:"quiz_text"`
	SampleAnswer string `db:"sample_answer"`
}

type ProfileQuestionRepository struct {
	db IDatabase
}

func (pqr *ProfileQuestionRepository) FetchByQuestionID(questionID uint) (*model.ProfileQuestion, error) {
	dbQuestion := DBQuestionRow{}
	if err := pqr.db.QueryRow("Master", "SELECT * FROM ProfileQuestion WHERE user_id = :user_id", DBQuestionRow{
		QuestionID: int(questionID),
	}).StructScan(&dbQuestion); err != nil {
		return nil, err
	}
	return model.NewProfileQuestion(
		uint(dbQuestion.QuestionID),
		dbQuestion.QuestionText,
		dbQuestion.QuizText,
		dbQuestion.SampleAnswer,
	)
}

func (pqr *ProfileQuestionRepository) FetchAllQuestions() ([]model.ProfileQuestion, error) {
	var n int
	if err := pqr.db.QueryRow("Master", "SELECT COUNT(*) FROM ProfileQuestion").Scan(&n); err != nil {
		return nil, err
	}
	questions := make([]model.ProfileQuestion, 0, n)
	rows, err := pqr.db.Query("Master", "SELECT * FROM ProfileQuestion")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		dbQuestion := DBQuestionRow{}
		if err := rows.StructScan(&dbQuestion); err != nil {
			return nil, err
		}
		question, err := model.NewProfileQuestion(
			uint(dbQuestion.QuestionID),
			dbQuestion.QuestionText,
			dbQuestion.QuizText,
			dbQuestion.SampleAnswer,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, *question)
	}

	return questions, nil
}

func NewProfileQuestionRepository(db IDatabase) *ProfileQuestionRepository {
	return &ProfileQuestionRepository{
		db: db,
	}
}
