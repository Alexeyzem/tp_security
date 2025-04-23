package core

type Request struct {
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	GetParams  map[string][]string `json:"params"`
	Headers    map[string][]string `json:"headers"`
	Cookies    map[string]string   `json:"cookies"`
	PostParams map[string][]string `json:"post_params"`
}
