package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const (
	maxNotes   = 1000
	maxTextLen = 280
)

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type Store struct {
	mu     sync.RWMutex
	nextID int
	notes  map[int]Note
}

func NewStore() *Store { return &Store{nextID: 1, notes: map[int]Note{}} }

func (s *Store) Add(text string) (Note, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.notes) >= maxNotes {
		return Note{}, false
	}
	n := Note{ID: s.nextID, Text: text}
	s.notes[n.ID] = n
	s.nextID++
	return n, true
}

func (s *Store) Get(id int) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	return n, ok
}

func (s *Store) List() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Note, 0, len(s.notes))
	for id := 1; id < s.nextID; id++ {
		if n, ok := s.notes[id]; ok {
			out = append(out, n)
		}
	}
	return out
}

func RegisterRoutes(mux *http.ServeMux, s *Store) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": os.Getenv("APP_VERSION")})
	})

	mux.HandleFunc("GET /api/v1/notes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List())
	})

	mux.HandleFunc("POST /api/v1/notes", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil || in.Text == "" || len(in.Text) > maxTextLen {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "field text wajib diisi, maksimal 280 karakter"})
			return
		}
		n, ok := s.Add(in.Text)
		if !ok {
			writeJSON(w, http.StatusInsufficientStorage, map[string]string{"error": "batas jumlah catatan tercapai"})
			return
		}
		writeJSON(w, http.StatusCreated, n)
	})

	mux.HandleFunc("GET /api/v1/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id harus angka"})
			return
		}
		n, ok := s.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "catatan tidak ditemukan"})
			return
		}
		writeJSON(w, http.StatusOK, n)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
