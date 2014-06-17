package app

import (
	"appengine"
	"appengine/user"
	"errors"
	"github.com/MiniProfiler/go/miniprofiler"
	mp "github.com/MiniProfiler/go/miniprofiler_gae"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

var (
	Router      *mux.Router
	viewCache   map[string]*template.Template
	isDevServer bool
)

func init() {
	Router = mux.NewRouter()

	parseAndCacheTemplates()

	Router.NotFoundHandler = NewHandler(NotFound)
	Router.Handle("/", NewHandler(Index))

	http.Handle("/", Router)

	isDevServer = appengine.IsDevAppServer()

	miniprofiler.Position = "right"
}

func Index(c mp.Context, w http.ResponseWriter, r *http.Request) error {
	return Render(w, NewIncludes(c, w, r), "index")
}

func NewHandler(f func(mp.Context, http.ResponseWriter, *http.Request) error) http.Handler {
	return mp.NewHandler(func(c mp.Context, w http.ResponseWriter, r *http.Request) {
		err := f(c, w, r)
		if err != nil {
			// let the appengine handle it
			panic(err)
		}

		return
	})
}

func NotFound(c mp.Context, w http.ResponseWriter, r *http.Request) error {
	return Render(w, NewIncludes(c, w, r), "404")
}

func Render(w http.ResponseWriter, data interface{}, name string) error {
	t := viewCache[name]
	if t == nil {
		return errors.New("cannot find the view template [" + name + "].")
	}

	if err := t.Execute(w, data); err != nil {
		return err
	}

	return nil
}

const layout = "views/layout.html"

func parseAndCacheTemplates() {
	views := []string{
		"index",
		"404",
	}
	viewCache = make(map[string]*template.Template, len(views))

	for _, v := range views {
		t := template.Must(template.New("layout.html").ParseFiles("views/"+v+".html", layout))
		viewCache[v] = t
	}
}

//
// includes
//
type Includes struct {
	MiniProfiler template.HTML
	IsDev        bool
	IsAdmin      bool
}

func NewIncludes(c mp.Context, w http.ResponseWriter, r *http.Request) *Includes {
	i := &Includes{
		MiniProfiler: c.Includes(),
		IsDev:        isDevServer,
	}

	if cu := user.Current(c); cu != nil {
		i.IsAdmin = cu.Admin
	}

	return i
}
