package dashboard

import (
	"bufio"
	"net/http"
	"os"
	"time"

	"goction/internal/api/dashboard/templates"
	"goction/internal/config"
	"goction/internal/stats"
	"goction/internal/viewmodels"

	"github.com/gorilla/sessions"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func getSystemStats() viewmodels.SystemStats {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	uptime, _ := host.Uptime()

	return viewmodels.SystemStats{
		CPUUsage:    cpuPercent[0],
		MemoryUsage: memInfo.UsedPercent,
		Uptime:      time.Duration(uptime) * time.Second,
	}
}

func LoginHandler(cfg *config.Config, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "goction-dashboard")

		if r.Method == "POST" {
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == cfg.DashboardUsername && password == cfg.DashboardPassword {
				session.Values["authenticated"] = true
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		templates.WriteLogin(w, config.GoctionVersion)
	}
}

func LogoutHandler(store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "goction-dashboard")
		session.Values["authenticated"] = false
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getRecentLogs(logFilePath string, numLines int) ([]string, error) {
    file, err := os.Open(logFilePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > numLines {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func DashboardHandler(cfg *config.Config, statsManager *stats.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allStats := statsManager.GetAllStats()
		history := statsManager.GetAllHistory() // Utilisation de la nouvelle méthode

		recentLogs, err := getRecentLogs(cfg.LogFile, 50) // Get last 50 lines
		if err != nil {
			// Handle error, maybe log it
			recentLogs = []string{"Error reading logs: " + err.Error()}
		}

		data := viewmodels.DashboardData{
			Config:         cfg,
			Stats:          allStats,
			History:        history,
			RecentLogs:     recentLogs,
			GoctionVersion: config.GoctionVersion,
		}

		templates.WriteDashboard(w, data)
	}
}

func AuthMiddleware(store *sessions.CookieStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "goction-dashboard")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func SetupRoutes(router *http.ServeMux, cfg *config.Config, statsManager *stats.Manager, store *sessions.CookieStore) {
	router.HandleFunc("/login", LoginHandler(cfg, store))
	router.HandleFunc("/logout", LogoutHandler(store))
	router.HandleFunc("/", AuthMiddleware(store, DashboardHandler(cfg, statsManager)))
}
