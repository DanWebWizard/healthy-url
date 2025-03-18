package models

type UrlHealth struct {
	URL            string
	HealthyUrl     bool
	UnhealthyUrl   bool
	UnreachableUrl bool
	TimeTaken      int
	StatusCode     int
}

type URLStatus int

const (
	Healthy URLStatus = iota
	Unhealthy
	Unreachable
)
