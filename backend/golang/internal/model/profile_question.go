package model

type ProfileQuestion struct {
	questionID   uint
	questionText string
	quizText     string
	sampleAnswer string
}

func (pq *ProfileQuestion) GetQuestionID() uint {
	return pq.questionID
}

func (pq *ProfileQuestion) GetQuestionText() string {
	return pq.questionText
}

func (pq *ProfileQuestion) GetQuizText() string {
	return pq.quizText
}

func (pq *ProfileQuestion) GetSampleAnswer() string {
	return pq.sampleAnswer
}

func NewProfileQuestion(questionID uint, questionText string, quizText string, sampleAnswer string) (*ProfileQuestion, error) {
	return &ProfileQuestion{
		questionID:   questionID,
		questionText: questionText,
		quizText:     quizText,
		sampleAnswer: sampleAnswer,
	}, nil
}
