package quiz

/*
	Question represents each the individual elements of a quiz. It stores the users answer and whether that answer
	is correct.
*/
type Question struct {
	QuestionText string
	Answer       string
	UserAnswer   string
	Correct      *bool
}

/*
	Creates a new question using the given questionText and answer.
	UserAnswer is initially stored as "<empty>" and the Correct pointer is initially stored as nil.
	These values are updated when a quiz is taken.
*/
func NewQuestion(ques string, ans string) Question {
	question := Question{ques, ans, "<empty>", nil}
	return question
}

/*
	Returns a string representation of a Question
*/
func (q Question) String() string {
	var resultBox string
	if q.Correct == nil {
		resultBox = "[   ]"
	} else if *q.Correct {
		resultBox = "[ √ ]"
	} else {
		resultBox = "[ X ]"
	}

	return q.QuestionText + "\t" + q.Answer + "\t" + q.UserAnswer + "\t" + resultBox
}
