package trust

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Peer struct {
	FP        string    `json:"fp"`
	Label     string    `json:"label"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type Store struct {
	path string
	mu   sync.Mutex
	m    map[string]Peer
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{
		path: filepath.Join(dir, "trust.json"),
		m:    map[string]Peer{},
	}
	_ = s.load()
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var st struct {
		Peers []Peer `json:"peers"`
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return err
	}
	for _, p := range st.Peers {
		s.m[p.FP] = p
	}
	return nil
}

func (s *Store) save() error {
	st := struct {
		Peers []Peer `json:"peers"`
	}{Peers: make([]Peer, 0, len(s.m))}
	for _, p := range s.m {
		st.Peers = append(st.Peers, p)
	}
	b, err := json.MarshalIndent(&st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Get(fp string) (Peer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[fp]
	return p, ok
}

func (s *Store) Add(fp, label string) error {
	if fp == "" {
		return errors.New("empty fp")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	p := s.m[fp]
	if p.FP == "" {
		p = Peer{FP: fp, Label: label, FirstSeen: now}
	}
	p.LastSeen = now
	s.m[fp] = p
	return s.save()
}

func (s *Store) Remove(fp string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, fp)
	return s.save()
}
