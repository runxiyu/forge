package web

import (
        "net/http"
        "net/url"
        "sort"
        "strings"
)

type (
        Vars        map[string]string
        HandlerFunc func(http.ResponseWriter, *http.Request, Vars)
)

type UserResolver func(*http.Request) (id int, username string, err error)

type ErrorRenderers struct {
        BadRequest      func(http.ResponseWriter, *BaseData, string)
        BadRequestColon func(http.ResponseWriter, *BaseData)
        NotFound        func(http.ResponseWriter, *BaseData)
        ServerError     func(http.ResponseWriter, *BaseData, string)
}

type dirPolicy int

const (
	dirIgnore dirPolicy = iota
	dirRequire
	dirForbid
	dirRequireIfEmpty
)

type patKind uint8

const (
	lit patKind = iota
	param
	splat
	group // @group, must be first token
)

type patSeg struct {
	kind patKind
	lit  string
	key  string
}

type route struct {
	method     string
	rawPattern string
	wantDir    dirPolicy
	ifEmptyKey string
	segs       []patSeg
	h          HandlerFunc
	hh         http.Handler
	priority   int
}

type Router struct {
	routes       []route
	errors       ErrorRenderers
	user         UserResolver
	global       any
	reverseProxy bool
}

func NewRouter() *Router { return &Router{} }

func (r *Router) Global(v any) *Router                { r.global = v; return r }
func (r *Router) ReverseProxy(enabled bool) *Router   { r.reverseProxy = enabled; return r }
func (r *Router) Errors(e ErrorRenderers) *Router     { r.errors = e; return r }
func (r *Router) UserResolver(u UserResolver) *Router { r.user = u; return r }

type RouteOption func(*route)

func WithDir() RouteOption    { return func(rt *route) { rt.wantDir = dirRequire } }
func WithoutDir() RouteOption { return func(rt *route) { rt.wantDir = dirForbid } }
func WithDirIfEmpty(param string) RouteOption {
	return func(rt *route) { rt.wantDir = dirRequireIfEmpty; rt.ifEmptyKey = param }
}

func (r *Router) GET(pattern string, f HandlerFunc, opts ...RouteOption) {
	r.handle("GET", pattern, f, nil, opts...)
}

func (r *Router) POST(pattern string, f HandlerFunc, opts ...RouteOption) {
	r.handle("POST", pattern, f, nil, opts...)
}

func (r *Router) ANY(pattern string, f HandlerFunc, opts ...RouteOption) {
	r.handle("", pattern, f, nil, opts...)
}

func (r *Router) ANYHTTP(pattern string, hh http.Handler, opts ...RouteOption) {
	r.handle("", pattern, nil, hh, opts...)
}

func (r *Router) handle(method, pattern string, f HandlerFunc, hh http.Handler, opts ...RouteOption) {
	want := dirIgnore
	if strings.HasSuffix(pattern, "/") {
		want = dirRequire
		pattern = strings.TrimSuffix(pattern, "/")
	} else if pattern != "" {
		want = dirForbid
	}
	segs, prio := compilePattern(pattern)
	rt := route{
		method:     method,
		rawPattern: pattern,
		wantDir:    want,
		segs:       segs,
		h:          f,
		hh:         hh,
		priority:   prio,
	}
	for _, o := range opts {
		o(&rt)
	}
	r.routes = append(r.routes, rt)

	sort.SliceStable(r.routes, func(i, j int) bool {
		return r.routes[i].priority > r.routes[j].priority
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
        segments, dirMode, err := splitAndUnescapePath(req.URL.EscapedPath())
        if err != nil {
                bd := &BaseData{Global: r.global}
                r.err400(w, bd, "Error parsing request URI: "+err.Error())
                return
        }
        for _, s := range segments {
                if strings.Contains(s, ":") {
                        bd := &BaseData{Global: r.global}
                        r.err400Colon(w, bd)
                        return
                }
        }

        bd := &BaseData{
                Global:      r.global,
                URLSegments: segments,
                DirMode:     dirMode,
        }
        if r.user != nil {
                uid, uname, uerr := r.user(req)
                if uerr != nil {
                        r.err500(w, bd, "Error getting user info from request: "+uerr.Error())
                        return
                }
                bd.UserID = uid
                bd.Username = uname
        }

        req = req.WithContext(WithBaseData(req.Context(), bd))

        method := req.Method

        for _, rt := range r.routes {
                if rt.method != "" &&
                        !(rt.method == method || (method == http.MethodHead && rt.method == http.MethodGet)) {
                        continue
                }
                ok, vars, sepIdx := match(rt.segs, segments)
                if !ok {
                        continue
                }
                switch rt.wantDir {
                case dirRequire:
                        if !dirMode && redirectAddSlash(w, req) {
                                return
                        }
                case dirForbid:
                        if dirMode && redirectDropSlash(w, req) {
                                return
                        }
                case dirRequireIfEmpty:
                        if v, _ := vars[rt.ifEmptyKey]; v == "" && !dirMode && redirectAddSlash(w, req) {
                                return
                        }
                }

                if g, ok := vars["group"]; ok {
                        if g == "" {
                                bd.GroupPath = []string{}
                        } else {
                                bd.GroupPath = strings.Split(g, "/")
                        }
                        delete(vars, "group")
                }
                bd.SeparatorIndex = sepIdx

                req = req.WithContext(WithBaseData(req.Context(), bd))

                if rt.h != nil {
                        rt.h(w, req, vars)
                } else if rt.hh != nil {
                        rt.hh.ServeHTTP(w, req)
                } else {
                        r.err500(w, bd, "route has no handler")
                }
                return
        }
        r.err404(w, bd)
}

func compilePattern(pat string) ([]patSeg, int) {
	if pat == "" || pat == "/" {
		return nil, 1000
	}
	pat = strings.Trim(pat, "/")
	raw := strings.Split(pat, "/")

	segs := make([]patSeg, 0, len(raw))
	prio := 0
	for i, t := range raw {
		switch {
		case t == "@group":
			if i != 0 {
				segs = append(segs, patSeg{kind: lit, lit: t})
				prio += 10
				continue
			}
			segs = append(segs, patSeg{kind: group})
			prio += 1
		case strings.HasPrefix(t, ":"):
			segs = append(segs, patSeg{kind: param, key: t[1:]})
			prio += 5
		case strings.HasPrefix(t, "*"):
			segs = append(segs, patSeg{kind: splat, key: t[1:]})
		default:
			segs = append(segs, patSeg{kind: lit, lit: t})
			prio += 10
		}
	}
	return segs, prio
}

func match(pat []patSeg, segs []string) (bool, map[string]string, int) {
	vars := make(map[string]string)
	i := 0
	sepIdx := -1
	for pi := 0; pi < len(pat); pi++ {
		ps := pat[pi]
		switch ps.kind {
		case group:
			start := i
			for i < len(segs) && segs[i] != "-" {
				i++
			}
			if start < i {
				vars["group"] = strings.Join(segs[start:i], "/")
			} else {
				vars["group"] = ""
			}
			if i < len(segs) && segs[i] == "-" {
				sepIdx = i
			}
		case lit:
			if i >= len(segs) || segs[i] != ps.lit {
				return false, nil, -1
			}
			i++
		case param:
			if i >= len(segs) {
				return false, nil, -1
			}
			vars[ps.key] = segs[i]
			i++
		case splat:
			if i < len(segs) {
				vars[ps.key] = strings.Join(segs[i:], "/")
				i = len(segs)
			} else {
				vars[ps.key] = ""
			}
			pi = len(pat)
		}
	}
	if i != len(segs) {
		return false, nil, -1
	}
	return true, vars, sepIdx
}

func splitAndUnescapePath(escaped string) ([]string, bool, error) {
	if escaped == "" {
		return nil, false, nil
	}
	dir := strings.HasSuffix(escaped, "/")
	path := strings.Trim(escaped, "/")
	if path == "" {
		return []string{}, dir, nil
	}
	raw := strings.Split(path, "/")
	out := make([]string, 0, len(raw))
	for _, seg := range raw {
		u, err := url.PathUnescape(seg)
		if err != nil {
			return nil, dir, err
		}
		if u != "" {
			out = append(out, u)
		}
	}
	return out, dir, nil
}

func redirectAddSlash(w http.ResponseWriter, r *http.Request) bool {
	u := *r.URL
	u.Path = u.EscapedPath() + "/"
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	return true
}

func redirectDropSlash(w http.ResponseWriter, r *http.Request) bool {
	u := *r.URL
	u.Path = strings.TrimRight(u.EscapedPath(), "/")
	if u.Path == "" {
		u.Path = "/"
	}
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	return true
}

func (r *Router) err400(w http.ResponseWriter, b *BaseData, msg string) {
        if r.errors.BadRequest != nil {
                r.errors.BadRequest(w, b, msg)
                return
        }
        http.Error(w, msg, http.StatusBadRequest)
}

func (r *Router) err400Colon(w http.ResponseWriter, b *BaseData) {
        if r.errors.BadRequestColon != nil {
                r.errors.BadRequestColon(w, b)
                return
        }
        http.Error(w, "bad request", http.StatusBadRequest)
}

func (r *Router) err404(w http.ResponseWriter, b *BaseData) {
        if r.errors.NotFound != nil {
                r.errors.NotFound(w, b)
                return
        }
        http.NotFound(w, nil)
}

func (r *Router) err500(w http.ResponseWriter, b *BaseData, msg string) {
        if r.errors.ServerError != nil {
                r.errors.ServerError(w, b, msg)
                return
        }
        http.Error(w, msg, http.StatusInternalServerError)
}
