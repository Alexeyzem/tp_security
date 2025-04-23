package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/tp_security/internal/core"
)

const maxGoroutine = 1000

type repo interface {
	GetOne(context.Context, int) (req string, resp string, err error)
	GetAll(context.Context) (map[int][2]string, error)
	Save(ctx context.Context, request, response string) error
}

type Contr struct {
	repo          repo
	scanningPaths []string
}

func New(repo repo, paths []string) *Contr {
	return &Contr{repo, paths}
}

func (c *Contr) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, resp, err := c.repo.GetOne(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req == "" {
		http.Error(w, "wrong request id", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

	res, err := json.Marshal(
		map[string]string{
			"Request":  req,
			"Response": resp,
		},
	)

	_, err = w.Write(res)
	if err != nil {
		log.Println(err)
	}
}

func (c *Contr) Get(w http.ResponseWriter, r *http.Request) {
	dataMap, err := c.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	var res []struct {
		ID       int    `json:"id"`
		Request  string `json:"request"`
		Response string `json:"response"`
	}

	for k, v := range dataMap {
		res = append(
			res, struct {
				ID       int    `json:"id"`
				Request  string `json:"request"`
				Response string `json:"response"`
			}{ID: k, Request: v[0], Response: v[1]},
		)
	}

	resp, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println(err)
	}
}

func (c *Contr) Repeat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, _, err := c.repo.GetOne(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req == "" {
		http.Error(w, "wrong request id", http.StatusNotFound)
		return
	}

	requestStruct := &core.Request{}
	err = json.Unmarshal([]byte(req), requestStruct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := c.sendReq(requestStruct, 5*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	saveResp, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	err = c.repo.Save(r.Context(), req, string(saveResp))
	if err != nil {
		log.Println(err)
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Println(err)
	}
}

func (c *Contr) Scan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, _, err := c.repo.GetOne(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req == "" {
		http.Error(w, "wrong request id", http.StatusNotFound)
		return
	}

	requestStruct := &core.Request{}
	err = json.Unmarshal([]byte(req), requestStruct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var dangerousPath []struct {
		Path string
		Resp *core.Response
	}
	wg := sync.WaitGroup{}
	wg.Add(len(c.scanningPaths))
	pool := make(chan struct{}, maxGoroutine)

	for _, v := range c.scanningPaths {
		pool <- struct{}{}
		requestStruct.Path = v
		go func() {
			defer wg.Done()
			defer func() { <-pool }()
			resp, err := c.sendReq(requestStruct, time.Second)
			if err != nil {
				log.Println(err)
			}

			if resp != nil && resp.Code != http.StatusNotFound {
				dangerousPath = append(
					dangerousPath, struct {
						Path string
						Resp *core.Response
					}{Path: v, Resp: resp},
				)
			}
		}()
	}

	wg.Wait()
	w.WriteHeader(http.StatusOK)
	if len(dangerousPath) > 0 {
		res, err := json.Marshal(
			map[string][]struct {
				Path string
				Resp *core.Response
			}{
				"Paths": dangerousPath,
			},
		)
		_, err = w.Write(res)
		if err != nil {
			log.Println(err)
		}
	} else {
		_, err = w.Write([]byte("No dangerous paths"))
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Contr) sendReq(req *core.Request, waitTime time.Duration) (*core.Response, error) {
	getParams := url.Values{}
	for k, v := range req.GetParams {
		for _, value := range v {
			getParams.Add(k, value)
		}
	}

	u := &url.URL{
		Scheme:   "http",
		Host:     req.Host,
		Path:     req.Path,
		RawQuery: getParams.Encode(),
	}

	var request *http.Request
	var err error
	var body io.Reader

	if req.Method == http.MethodPost && len(req.PostParams) > 0 {
		body = bytes.NewBufferString(url.Values(req.PostParams).Encode())
		request, err = http.NewRequest(req.Method, u.String(), body)
	} else {
		request, err = http.NewRequest(req.Method, u.String(), nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for k, vals := range req.Headers {
		request.Header[http.CanonicalHeaderKey(k)] = vals
	}

	for name, value := range req.Cookies {
		request.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	if req.Method == http.MethodPost && body != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	request.Host = req.Host

	targetConn, err := net.DialTimeout("tcp", req.Host, waitTime)
	if err != nil {
		return nil, err
	}

	if err := request.Write(targetConn); err != nil {
		return nil, fmt.Errorf("failed to write request: %v", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), request)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	defer resp.Body.Close()

	response := &core.Response{
		Code:    resp.StatusCode,
		Headers: toMapStringSlice(resp.Header),
		Message: resp.Status,
	}

	var bodyResp []byte
	_, err = resp.Body.Read(bodyResp)
	if err != nil {
		log.Println("error reading response body:", err)
	}
	response.Body = string(bodyResp)

	return response, nil
}

func toMapStringSlice[IN ~map[string][]string](in IN) map[string][]string {
	res := make(map[string][]string, len(in))
	for k, v := range in {
		for _, value := range v {
			res[k] = append(res[k], value)
		}
	}

	return res
}
