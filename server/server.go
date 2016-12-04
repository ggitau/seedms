package server

import (
	"errors"
	"golang.org/x/net/context"
	"github.com/tomogoma/seedms/server/proto"
	"net/http"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/server/helper"
)

type Logger interface {
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

type TokenValidator interface {
	Validate(token string) (*token.Token, error)
	IsClientError(error) bool
}

type Model interface {
}

type Server struct {
	token TokenValidator
	model Model
	log   Logger
	tIDCh chan int
	id    string
}

const (
	SomethingWickedError = "Something wicked happened"
)

var ErrorNilTokenValidator = errors.New("TokenValidator was nil")
var ErrorNilRiderModel = errors.New("Model was nil");
var ErrorNilLogger = errors.New("Logger was nil");

func New(id string, m Model, tv TokenValidator, lg Logger) (*Server, error) {
	if m == nil {
		return nil, ErrorNilRiderModel
	}
	if tv == nil {
		return nil, ErrorNilTokenValidator
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{id: id, model:m, token: tv, log:lg, tIDCh: tIDCh}, nil
}

func (s *Server) Hello(c context.Context, req *seed.HelloRequest, resp *seed.HelloResponse) error {
	resp.Id = s.id
	tID := <-s.tIDCh
	s.log.Info("%d - Hello request", tID)
	_, err := s.token.Validate(req.Token);
	if err != nil {
		if s.token.IsClientError(err) {
			resp.Code = http.StatusUnauthorized
			resp.Detail = err.Error()
			s.log.Info("%d - Token was invalid", tID)
			return nil
		}
		resp.Code = http.StatusInternalServerError
		resp.Detail = SomethingWickedError
		s.log.Info("%d - Failed to validate user token: %s", tID, err)
		return nil
	}
	resp.Code = http.StatusOK
	resp.Greeting = "Hello " + req.Name
	s.log.Info("%d - complete", tID)
	return nil
}
