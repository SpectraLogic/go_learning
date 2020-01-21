package quiz

// Question represents each the individual elements of a quiz. It stores the users answer and whether that answer
// is correct.
type Question struct {
	QuestionText string
	Answer       string
	UserAnswer   string
	Correct      *bool
}

// String returns a string representation of a Question
func (q Question) String() string {
	var resultBox string
	switch {
	case q.Correct == nil:
		resultBox = "[   ]"
	case *q.Correct:
		resultBox = "[ √ ]"
	default:
		resultBox = "[ X ]"
	}

	return q.QuestionText + "\t" + q.Answer + "\t" + q.UserAnswer + "\t" + resultBox
}
