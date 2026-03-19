package repository

import (
	"fmt"
	"sync"

	"github.com/logitrack/core/internal/model"
)

type RouteRepository interface {
	Create(route model.Route) (model.Route, error)
	Update(route model.Route) error
	GetByDriverAndDate(driverID, date string) (model.Route, error)
	GetByID(id string) (model.Route, error)
}

type inMemoryRouteRepository struct {
	mu     sync.RWMutex
	routes []model.Route
}

func NewInMemoryRouteRepository() RouteRepository {
	return &inMemoryRouteRepository{}
}

func (r *inMemoryRouteRepository) Create(route model.Route) (model.Route, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, route)
	return route, nil
}

func (r *inMemoryRouteRepository) Update(route model.Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rt := range r.routes {
		if rt.ID == route.ID {
			r.routes[i] = route
			return nil
		}
	}
	return fmt.Errorf("route not found")
}

func (r *inMemoryRouteRepository) GetByDriverAndDate(driverID, date string) (model.Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, route := range r.routes {
		if route.DriverID == driverID && route.Date == date {
			return route, nil
		}
	}
	return model.Route{}, fmt.Errorf("no route found for driver %s on %s", driverID, date)
}

func (r *inMemoryRouteRepository) GetByID(id string) (model.Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, route := range r.routes {
		if route.ID == id {
			return route, nil
		}
	}
	return model.Route{}, fmt.Errorf("route not found")
}
