package runtimeapp

import (
	"localShareGo/internal/auth"
	"localShareGo/internal/clipboard"
	"localShareGo/internal/config"
	"localShareGo/internal/httpserver"
	"localShareGo/internal/network"
	"localShareGo/internal/store"
)

type RouteOverview struct {
	NaiveDesktop string `json:"naiveDesktop"`
	Web          string `json:"web"`
}

type ServiceOverview struct {
	Clipboard   clipboard.ClipboardStatus   `json:"clipboard"`
	HTTPServer  httpserver.HttpServerStatus `json:"httpServer"`
	Auth        auth.AuthStatus             `json:"auth"`
	Session     auth.SessionSnapshot        `json:"session"`
	Persistence store.PersistenceStatus     `json:"persistence"`
	Network     network.NetworkStatus       `json:"network"`
}

type AppBootstrap struct {
	AppName       string               `json:"appName"`
	Routes        RouteOverview        `json:"routes"`
	RuntimeConfig config.RuntimeConfig `json:"runtimeConfig"`
	Paths         config.AppPaths      `json:"paths"`
	Services      ServiceOverview      `json:"services"`
}

type ConnectivityCheck struct {
	Host           string  `json:"host"`
	URL            string  `json:"url"`
	TCPOk          bool    `json:"tcpOk"`
	HTTPOk         bool    `json:"httpOk"`
	HTTPStatusLine *string `json:"httpStatusLine"`
	Error          *string `json:"error"`
}

type ConnectivityReport struct {
	BindHost      string              `json:"bindHost"`
	PreferredPort int                 `json:"preferredPort"`
	EffectivePort int                 `json:"effectivePort"`
	ServerState   string              `json:"serverState"`
	ServerError   *string             `json:"serverError"`
	AccessURL     string              `json:"accessUrl"`
	Checks        []ConnectivityCheck `json:"checks"`
}
