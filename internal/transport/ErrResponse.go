package transport

const (
	NOT_FOUND   = "NOT_FOUND"
	BAD_REQUEST = "BAD_REQUEST"
)

type ErrResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
