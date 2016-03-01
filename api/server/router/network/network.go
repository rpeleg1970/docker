package network

import (
	"github.com/docker/docker/api/server/router"
	"sync"
)

// nameLocker allows fine grained locking on network names
type nameLocker struct {
	mutex  sync.Mutex
	muxMap map[string]*sync.Mutex
}

// networkRouter is a router to talk with the network controller
type networkRouter struct {
	backend  Backend
	routes   []router.Route
	netNames nameLocker
}

// NewRouter initializes a new network router
func NewRouter(b Backend) router.Router {
	r := &networkRouter{
		backend: b,
	}
	r.initRoutes()
	r.netNames.init()
	return r
}

// Routes returns the available routes to the network controller
func (r *networkRouter) Routes() []router.Route {
	return r.routes
}

func (r *networkRouter) initRoutes() {
	r.routes = []router.Route{
		// GET
		router.NewGetRoute("/networks", r.getNetworksList),
		router.NewGetRoute("/networks/{id:.*}", r.getNetwork),
		// POST
		router.NewPostRoute("/networks/create", r.postNetworkCreate),
		router.NewPostRoute("/networks/{id:.*}/connect", r.postNetworkConnect),
		router.NewPostRoute("/networks/{id:.*}/disconnect", r.postNetworkDisconnect),
		// DELETE
		router.NewDeleteRoute("/networks/{id:.*}", r.deleteNetwork),
	}
}

func (n *nameLocker) init() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.muxMap == nil {
		n.muxMap = make(map[string]*sync.Mutex)
	}
}

func (n *nameLocker) lock(name string) {
	n.mutex.Lock()
	if n.muxMap[name] == nil {
		n.muxMap[name] = &sync.Mutex{}
	}
	n.mutex.Unlock()
	n.muxMap[name].Lock()
}

func (n *nameLocker) unlock(name string) {
	n.mutex.Lock()
	if n.muxMap[name] != nil {
		n.muxMap[name].Unlock()
		if (*n.muxMap[name] == sync.Mutex{}) { // indicating no locks
			delete(n.muxMap, name)
		}
	}
	n.mutex.Unlock()
}
