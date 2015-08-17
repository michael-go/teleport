package session

import (
	"time"

	"github.com/gravitational/teleport/backend"
)

type SessionServer interface {
	GetSessions() ([]Session, error)
	GetSession(id string) (*Session, error)
	DeleteSession(id string) error
	UpsertSession(id string, ttl time.Duration) error
	UpsertParty(id string, p Party, ttl time.Duration) error
}

type server struct {
	bk backend.JSONCodec
}

func New(bk backend.Backend) *server {
	return &server{
		bk: backend.JSONCodec{bk},
	}
}

func (s *server) GetSessions() ([]Session, error) {
	keys, err := s.bk.GetKeys([]string{"sessions"})
	if err != nil {
		return nil, err
	}
	out := []Session{}
	for _, sid := range keys {
		se, err := s.GetSession(sid)
		if backend.IsNotFound(err) {
			continue
		}
		out = append(out, *se)
	}
	return out, nil
}

func (s *server) GetSession(id string) (*Session, error) {
	if _, err := s.bk.GetVal([]string{"sessions", id}, "val"); err != nil {
		return nil, err
	}

	parties, err := s.bk.GetKeys([]string{"sessions", id, "parties"})

	if err != nil {
		return nil, err
	}
	out := []Party{}
	for _, pk := range parties {
		var p *Party
		err := s.bk.GetJSONVal([]string{"sessions", id, "parties"}, pk, &p)
		if err != nil {
			if backend.IsNotFound(err) { // key was expired
				continue
			}
			return nil, err
		}
		out = append(out, *p)
	}
	return &Session{ID: id, Parties: out}, nil
}

func (s *server) UpsertSession(id string, ttl time.Duration) error {
	return s.bk.UpsertVal([]string{"sessions", id}, "val", []byte("val"), ttl)
}

func (s *server) UpsertParty(id string, p Party, ttl time.Duration) error {
	if err := s.UpsertSession(id, ttl); err != nil {
		return err
	}
	return s.bk.UpsertJSONVal([]string{"sessions", id, "parties"}, p.ID, p, ttl)
}

func (s *server) DeleteSession(id string) error {
	return s.bk.DeleteBucket([]string{"sessions"}, id)
}

type Session struct {
	ID      string  `json:"id"`
	Parties []Party `json:"parties"`
}

type Party struct {
	ID         string    `json:"id"`
	Site       string    `json:"site"`
	User       string    `json:"user"`
	ServerAddr string    `json:"server_addr"`
	LastActive time.Time `json:"last_active"`
}