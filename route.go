package main

import (
	"DemaeDominos/dominos"
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"net/http"
	"strings"
)

type Route struct {
	Actions []Action
}

// Action contains information about how a specified action should be handled.
type Action struct {
	ActionName  string
	Callback    func(*Response)
	XMLType     XMLType
	ServiceType string
}

func NewRoute() Route {
	return Route{}
}

// RoutingGroup defines a group of actions for a given service type.
type RoutingGroup struct {
	Route       *Route
	ServiceType string
}

// HandleGroup returns a routing group type for the given service type.
func (r *Route) HandleGroup(serviceType string) RoutingGroup {
	return RoutingGroup{
		Route:       r,
		ServiceType: serviceType,
	}
}

func (r *RoutingGroup) NormalResponse(action string, function func(*Response)) {
	r.Route.Actions = append(r.Route.Actions, Action{
		ActionName:  action,
		Callback:    function,
		XMLType:     Normal,
		ServiceType: r.ServiceType,
	})
}

func (r *RoutingGroup) MultipleRootNodes(action string, function func(*Response)) {
	r.Route.Actions = append(r.Route.Actions, Action{
		ActionName:  action,
		Callback:    function,
		XMLType:     MultipleRootNodes,
		ServiceType: r.ServiceType,
	})
}

func (r *Route) Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var temp string
		row := pool.QueryRow(context.Background(), QueryUserBasket, req.Header.Get("X-WiiID"))
		err := row.Scan(&temp)
		if err != nil {
			// There are 2 possible reasons for this happening.
			// 1. Postgres spontaneously died in which case we will be alerted else where
			// 2. An unauthorized user bypassed the boot lock and tried to access.
			// We deny entry and report.
			printError(w, "Unauthorized user", http.StatusUnauthorized)
			_, ok := req.Header["X-Address"]
			if ok {
				// This request was made from a Wii.
				// If it wasn't ignore because it could be web crawlers among other things.
				PostDiscordWebhook("Unauthorized user!", fmt.Sprintf("A user who is not registered in the database has attempted to access the channel! Wii ID: %s", req.Header.Get("X-WiiID")), config.ErrorWebhook, 16711711)
			}

			_ = dataDog.Incr("demae-dominos.unauthorized_users", nil, 1)
			return
		}

		// First check if it is an image route.
		if strings.Contains(req.URL.Path, "itemimg") {
			splitUrl := strings.Split(req.URL.Path, "/")
			imageName := splitUrl[len(splitUrl)-1]
			dom, err := dominos.NewDominos(pool, req)
			if err != nil {
				// Most likely the user is not registered.
				printError(w, err.Error(), http.StatusUnauthorized)
				sentry.CaptureException(err)
				return
			}

			img := dom.DownloadAndReturnImage(imageName)
			w.Write(img)
			return
		} else if strings.Contains(req.URL.Path, "logoimg2") {
			// Serve Domino's logo
			w.Write(dominos.DominosLogo)
			return
		}

		// If this is a POST request it is either an actual request or an error.
		var actionName string
		var serviceType string
		var userAgent string
		if req.Method == "POST" {
			req.ParseForm()
			actionName = req.PostForm.Get("action")
			userAgent = req.PostForm.Get("platform")
			serviceType = "nwapi.php"
		} else {
			actionName = req.URL.Query().Get("action")
			userAgent = req.URL.Query().Get("platform")
			serviceType = strings.Replace(req.URL.Path, "/", "", -1)
		}

		if userAgent != "wii" {
			printError(w, "Invalid request.", http.StatusBadRequest)
			return
		}

		// Ensure we can route to this action before processing.
		// Search all registered actions and find a matching action.
		var action Action
		for _, routeAction := range r.Actions {
			if routeAction.ActionName == actionName && routeAction.ServiceType == serviceType {
				action = routeAction
			}
		}

		// Action is only properly populated if we found it previously.
		if action.ActionName == "" && action.ServiceType == "" {
			printError(w, "Unknown action was passed.", http.StatusBadRequest)
			return
		}

		resp := NewResponse(req, &w, action.XMLType)
		action.Callback(resp)

		if resp.hasError {
			// Response was already written by callback function.
			return
		}

		contents, err := resp.toXML()
		if err != nil {
			printError(w, err.Error(), http.StatusInternalServerError)
			sentry.CaptureException(err)
			return
		}

		w.Write([]byte(contents))
	})
}
