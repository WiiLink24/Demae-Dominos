package main

import (
	"DemaeDominos/dominos"
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
		fmt.Println(req.URL.String())
		// First check if it is an image route.
		if strings.Contains(req.URL.Path, "itemimg") {
			splitUrl := strings.Split(req.URL.Path, "/")
			imageName := splitUrl[len(splitUrl)-1]
			dom, err := dominos.NewDominos(req)
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
		if resp == nil {
			// Already wrote the error in the caller.
			return
		}

		action.Callback(resp)

		contents, err := resp.toXML()
		if err != nil {
			printError(w, err.Error(), http.StatusInternalServerError)
			sentry.CaptureException(err)
			return
		}

		w.Write([]byte(contents))
	})
}
