package model

import (
	"time"
)

type Package struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Readme       string     `json:"readme"`
	Releases     []*Release `json:"releases"`
	Author       string     `json:"author"`
	CreationDate time.Time  `json:"creationDate"`
	LastUpdated  time.Time  `json:"lastUpdated"`
	Rating       float64    `json:"rating"`
	Certified    bool       `json:"certified"`
	Downloads    int        `json:"downloads"`
}

type Release struct {
	ID          int       `json:"id"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Date        time.Time `json:"date"`
	Rating      float64   `json:"rating"`
	Manifests   []string  `json:"manifests"`
	PackageId   int       `json:"packageId"`
}

type Results struct {
	Count   int        `json:"count"`
	Offset  int        `json:"offset"`
	Total   int        `json:"total"`
	Results []*Package `json:"results"`
}
