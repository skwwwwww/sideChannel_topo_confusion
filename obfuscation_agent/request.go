package main

type DownstreamNodeConfig struct {
	DNS        string  `json:"DNS"`
	ServiceNum int     `json:"ServiceNum"`
	Rps        float64 `json:"Rps"`
	ErrorRate  float64 `json:"ErrorRate"`
}
