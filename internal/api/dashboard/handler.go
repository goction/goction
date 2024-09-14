package dashboard

import (
	"html/template"
	"net/http"
	"time"

	"goction/internal/config"
	"goction/internal/stats"

	"github.com/gorilla/sessions"
)

type Handler struct {
	config    *config.Config
	stats     *stats.Manager
	templates *template.Template
	store     *sessions.CookieStore
}

type GoctionStat struct {
	TotalCalls      int
	SuccessfulCalls int
	SuccessRate     float64
	LastExecuted    time.Time
}

type DashboardData struct {
	Goctions           []string
	Stats              map[string]GoctionStat
	TotalExecutions    int
	OverallSuccessRate float64
}

func NewHandler(cfg *config.Config, statsManager *stats.Manager) (*Handler, error) {
	templates, err := template.ParseGlob("internal/api/dashboard/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		config:    cfg,
		stats:     statsManager,
		templates: templates,
		store:     sessions.NewCookieStore([]byte(cfg.APIToken)),
	}, nil
}

func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	allStats := h.stats.GetAllStats() 
	data := DashboardData{
		Goctions: make([]string, 0, len(allStats)),
		Stats:    make(map[string]GoctionStat),
	}

	totalCalls := 0
	totalSuccessfulCalls := 0

	for name, stat := range allStats {
		data.Goctions = append(data.Goctions, name)
		successRate := 0.0
		if stat.TotalCalls > 0 {
			successRate = float64(stat.SuccessfulCalls) / float64(stat.TotalCalls) * 100
		}
		data.Stats[name] = GoctionStat{
			TotalCalls:      stat.TotalCalls,
			SuccessfulCalls: stat.SuccessfulCalls,
			SuccessRate:     successRate,
			LastExecuted:    stat.LastExecuted,
		}

		totalCalls += stat.TotalCalls
		totalSuccessfulCalls += stat.SuccessfulCalls
	}

	data.TotalExecutions = totalCalls
	if totalCalls > 0 {
		data.OverallSuccessRate = float64(totalSuccessfulCalls) / float64(totalCalls) * 100
	}

	h.templates.ExecuteTemplate(w, "dashboard.html", data)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == h.config.DashboardUsername && password == h.config.DashboardPassword {
			session, _ := h.store.Get(r, "dashboard-session")
			session.Values["authenticated"] = true
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	h.templates.ExecuteTemplate(w, "login.html", nil)
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "dashboard-session")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "dashboard-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
