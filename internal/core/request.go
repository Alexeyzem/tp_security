package core

type Request struct {
	Method     string              `json:"method"`
	Host       string              `json:"host"`
	Path       string              `json:"path"`
	GetParams  map[string][]string `json:"get_params"`
	Headers    map[string][]string `json:"headers"`
	Cookies    map[string]string   `json:"cookies"`
	PostParams map[string][]string `json:"post_params"`
}
