package main

import "time"

type NginxRaw struct {
	Time          time.Time `json:"time"`
	RemoteAddr    string    `json:"remote_addr"`
	Hostname      string    `json:"hostname"`
	Host          string    `json:"host"`
	Request       string    `json:"request"`
	Status        int       `json:"status"`
	BodyBytesSent int       `json:"body_bytes_sent"`
	HttpReferer   string    `json:"http_referer"`
	HttpUserAgent string    `json:"http_user_agent"`
	XSessionId    string    `json:"x_session_id"`
	XUserId       string    `json:"x_user_id"`
	XUserId2      string    `json:"x_user_id2"`
}

type Nginx struct {
	Time              time.Time `json:"time"`
	RemoteAddr        string    `json:"remote_addr"`
	Hostname          string    `json:"hostname"`
	Host              string    `json:"host"`
	CodUsuario        int       `json:"cod_usuario"`
	CodPersonificador int       `json:"cod_personificador"`
	Method            string    `json:"method"`
	Endpoint          string    `json:"endpoint"`
	QueryString       string    `json:"query_string"`
	Status            int       `json:"status"`
	BodyBytesSent     int       `json:"body_bytes_sent"`
	SessionId         string    `json:"session_id"`
	HttpUserAgent     string    `json:"http_user_agent"`
	HttpReferer       string    `json:"http_referer"`
	GeoStatus         string    `json:"geo_status"`
	CountryCode       string    `json:"country_code"`
	Region            string    `json:"region"`
	City              string    `json:"city"`
	Lat               float64   `json:"lat"`
	Lon               float64   `json:"lon"`
	As                string    `json:"as"`
	Mobile            bool      `json:"mobile"`
	Proxy             bool      `json:"proxy"`
	Hosting           bool      `json:"hosting"`
}

type NginxLine struct {
	Server            string    `json:"server"`
	File              string    `json:"file"`
	TimeLocal         time.Time `json:"time_local"`
	RemoteAddr        string    `json:"remote_addr"`
	CodUsuario        int       `json:"cod_usuario"`
	CodPersonificador int       `json:"cod_personificador"`
	Method            string    `json:"method"`
	Endpoint          string    `json:"endpoint"`
	QueryString       string    `json:"query_string"`
	Status            int       `json:"status"`
	BodyBytesSent     int       `json:"body_bytes_sent"`
	SessionId         string    `json:"session_id"`
	HttpUserAgent     string    `json:"http_user_agent"`
	CreatedAt         time.Time `json:"created_at"`
	HttpReferer       string    `json:"http_referer"`
	GeoStatus         string    `json:"geo_status"`
	CountryCode       string    `json:"country_code"`
	Region            string    `json:"region"`
	City              string    `json:"city"`
	Lat               float64   `json:"lat"`
	Lon               float64   `json:"lon"`
	As                string    `json:"as"`
	Mobile            bool      `json:"mobile"`
	Proxy             bool      `json:"proxy"`
	Hosting           bool      `json:"hosting"`
	Line              string    `json:"line"`
	Error             string    `json:"error"`
}
