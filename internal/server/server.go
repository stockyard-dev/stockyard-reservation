package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-reservation/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/reservations", s.listReservations)
	s.mux.HandleFunc("POST /api/reservations", s.createReservations)
	s.mux.HandleFunc("GET /api/reservations/export.csv", s.exportReservations)
	s.mux.HandleFunc("GET /api/reservations/{id}", s.getReservations)
	s.mux.HandleFunc("PUT /api/reservations/{id}", s.updateReservations)
	s.mux.HandleFunc("DELETE /api/reservations/{id}", s.delReservations)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listReservations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"reservations": oe(s.db.SearchReservations(q, filters))}); return }
	wj(w, 200, map[string]any{"reservations": oe(s.db.ListReservations())})
}

func (s *Server) createReservations(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Reservations
	json.NewDecoder(r.Body).Decode(&e)
	if e.GuestName == "" { we(w, 400, "guest_name required"); return }
	if e.Date == "" { we(w, 400, "date required"); return }
	if e.Time == "" { we(w, 400, "time required"); return }
	s.db.CreateReservations(&e)
	wj(w, 201, s.db.GetReservations(e.ID))
}

func (s *Server) getReservations(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetReservations(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateReservations(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetReservations(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Reservations
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.GuestName == "" { patch.GuestName = existing.GuestName }
	if patch.GuestPhone == "" { patch.GuestPhone = existing.GuestPhone }
	if patch.GuestEmail == "" { patch.GuestEmail = existing.GuestEmail }
	if patch.Date == "" { patch.Date = existing.Date }
	if patch.Time == "" { patch.Time = existing.Time }
	if patch.Table == "" { patch.Table = existing.Table }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateReservations(&patch)
	wj(w, 200, s.db.GetReservations(patch.ID))
}

func (s *Server) delReservations(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteReservations(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportReservations(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListReservations()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=reservations.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "guest_name", "guest_phone", "guest_email", "party_size", "date", "time", "table", "status", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.GuestName), fmt.Sprintf("%v", e.GuestPhone), fmt.Sprintf("%v", e.GuestEmail), fmt.Sprintf("%v", e.PartySize), fmt.Sprintf("%v", e.Date), fmt.Sprintf("%v", e.Time), fmt.Sprintf("%v", e.Table), fmt.Sprintf("%v", e.Status), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["reservations_total"] = s.db.CountReservations()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "reservation"}
	m["reservations"] = s.db.CountReservations()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}
