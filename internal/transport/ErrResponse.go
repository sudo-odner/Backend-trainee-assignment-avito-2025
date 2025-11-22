package transport

const (
	SERVER_ERROR = "SERVER_ERROR"
	NOT_FOUND    = "NOT_FOUND"
	BAD_REQUEST  = "BAD_REQUEST"
	TEAM_EXISTS  = "TEAM_EXISTS"
	PR_EXISTS    = "PR_EXISTS"
	PR_MERGED    = "PR_MERGED"
	NO_CANDIDATE = "NO_CANDIDATE"
	NOT_ASSIGNED = "NOT_ASSIGNED"
)

type ErrResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
