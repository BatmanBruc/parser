package queue

type WrapperMessage interface {
	Success()
	TryAgain()
	Reject()
	GetBody() []byte
}
