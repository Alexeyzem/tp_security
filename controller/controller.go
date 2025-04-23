package controller

import (
	"net/http"
)

type repo interface {
}

type Contr struct {
	repo repo
}

func New(repo repo) *Contr {
	return &Contr{repo}
}

func (c *Contr) GetOne(w http.ResponseWriter, r *http.Request) {}

func (c *Contr) Get(w http.ResponseWriter, r *http.Request) {}

func (c *Contr) Repeat(w http.ResponseWriter, r *http.Request) {}

func (c *Contr) Scan(w http.ResponseWriter, r *http.Request) {}
