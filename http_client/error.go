package http_client

var _ ReplyErr = (*replyErr)(nil)

// ReplyErr if resp.StatusCode != http.StatusOK, this enum will wrap resp data.
type ReplyErr interface {
	error
	StatusCode() int
	Body() []byte
}

type replyErr struct {
	err        error
	statusCode int
	body       []byte
}

func (r *replyErr) Error() string {
	return r.err.Error()
}

func (r *replyErr) StatusCode() int {
	return r.statusCode
}

func (r *replyErr) Body() []byte {
	return r.body
}

func newReplyErr(statusCode int, body []byte, err error) ReplyErr {
	return &replyErr{
		statusCode: statusCode,
		body:       body,
		err:        err,
	}
}

// IsReplyErr check if err is reply enum
func IsReplyErr(err error) (ReplyErr, bool) {
	if err == nil {
		return nil, false
	}

	e, ok := err.(ReplyErr)
	return e, ok
}
