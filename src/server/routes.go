package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"

	"github.com/lfkeitel/registry-webui/src/auth"
	"github.com/lfkeitel/registry-webui/src/controllers"
	mid "github.com/lfkeitel/registry-webui/src/server/middleware"
	"github.com/lfkeitel/registry-webui/src/utils"
)

func LoadRoutes(e *utils.Environment) http.Handler {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)

	r.Handler("GET", "/", midStack(e, http.HandlerFunc(rootHandler)))
	r.ServeFiles("/public/*filepath", http.Dir("./public"))

	authCont := controllers.NewAuthController(e)
	r.Handler("GET", "/login", midStack(e, http.HandlerFunc(authCont.Login)))
	r.Handler("POST", "/login", midStack(e, http.HandlerFunc(authCont.LoginPost)))
	r.Handler("GET", "/logout", midStack(e, http.HandlerFunc(authCont.Logout)))

	homeCont := controllers.NewHomeController(e)
	r.Handler("GET", "/home", midStackPlusAuth(e, http.HandlerFunc(homeCont.Home)))

	h := mid.Logging(e, r) // Logging
	h = mid.Panic(e, h)    // Panic catcher
	return h
}

func midStack(e *utils.Environment, h http.Handler) http.Handler {
	h = mid.SetSessionInfo(e, h) // Adds Environment and user information to request context
	h = context.ClearHandler(h)  // Clear Gorilla sessions
	return h
}

func midStackPlusAuth(e *utils.Environment, h http.Handler) http.Handler {
	h = mid.Auth(e, h)
	h = midStack(e, h)
	return h
}

// func devRouter(e *utils.Environment) http.Handler {
// 	r := httprouter.New()
// 	r.NotFound = http.HandlerFunc(notFoundHandler)

// 	devController := controllers.NewDevController(e)
// 	r.HandlerFunc("GET", "/dev/reloadtemp", devController.ReloadTemplates)
// 	r.HandlerFunc("GET", "/dev/reloadconf", devController.ReloadConfiguration)

// 	h := mid.CheckAdmin(r)
// 	h = mid.CheckAuth(h)
// 	return h
// }

// func debugRouter(e *utils.Environment) http.Handler {
// 	r := httprouter.New()
// 	r.NotFound = http.HandlerFunc(notFoundHandler)

// 	r.HandlerFunc("GET", "/debug/pprof", pprof.Index)
// 	r.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
// 	r.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
// 	r.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
// 	r.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)
// 	// Manually add support for paths linked to by index page at /debug/pprof/
// 	r.Handler("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
// 	r.Handler("GET", "/debug/pprof/heap", pprof.Handler("heap"))
// 	r.Handler("GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
// 	r.Handler("GET", "/debug/pprof/block", pprof.Handler("block"))

// 	r.HandlerFunc("GET", "/debug/heap-stats", heapStats)

// 	h := mid.CheckAdmin(r)
// 	h = mid.CheckAuth(h)
// 	return h
// }

// func heapStats(w http.ResponseWriter, r *http.Request) {
// 	var m runtime.MemStats
// 	runtime.ReadMemStats(&m)
// 	fmt.Fprintf(w,
// 		"HeapSys: %d, HeapAlloc: %d, HeapIdle: %d, HeapReleased: %d\n",
// 		m.HeapSys,
// 		m.HeapAlloc,
// 		m.HeapIdle,
// 		m.HeapReleased,
// 	)
// }

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		utils.NewEmptyAPIResponse().WriteResponse(w, http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
