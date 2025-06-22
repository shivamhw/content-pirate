package scrapper

type ScrErr string 

const (
	JobNotFound ScrErr = "JOB_NOT_FOUND"
)


func (err ScrErr) Error() string {
	return string(err)
}