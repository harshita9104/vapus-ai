package models

import (
	mpb "github.com/vapusdata-ecosystem/apis/protos/models/v1alpha1"
)

type APIAccessLogs struct {
	ID           int64  `json:"id" bun:"id,pk,autoincrement"`
	RequestedAt  int64  `json:"requestedAt" bun:"requested_at,notnull"`
	Endpoint     string `json:"endpoint" bun:"endpoint,notnull"`
	Error        bool   `json:"error,omitempty" bun:"error"`
	ErrorMessage string `json:"errorMessage,omitempty" bun:"error_message"`
	RemoteIP     string `json:"remoteIP" bun:"remote_ip,notnull"`
	UserAgent    string `json:"userAgent" bun:"user_agent,notnull"`
	StatusCode   int64  `json:"statusCode" bun:"status_code,notnull"`
	TotalLatency int64  `json:"totalLatency" bun:"total_latency,notnull"`
}

func (l *APIAccessLogs) ConvertFromPb(pb *mpb.APIAccessLog) *APIAccessLogs {
	if l == nil {
		return nil
	}

	obj := &APIAccessLogs{
		RequestedAt:  pb.RequestedAt,
		Endpoint:     pb.Endpoint,
		Error:        pb.Error,
		ErrorMessage: pb.ErrorMessage,
		RemoteIP:     pb.RemoteIp,
		UserAgent:    pb.UserAgent,
		StatusCode:   pb.StatusCode,
		TotalLatency: pb.TotalLatency,
	}

	return obj
}

func (l *APIAccessLogs) ConvertToPb() *mpb.APIAccessLog {
	if l == nil {
		return nil
	}

	obj := &mpb.APIAccessLog{
		RequestedAt:  l.RequestedAt,
		Endpoint:     l.Endpoint,
		Error:        l.Error,
		ErrorMessage: l.ErrorMessage,
		RemoteIp:     l.RemoteIP,
		UserAgent:    l.UserAgent,
		StatusCode:   l.StatusCode,
		TotalLatency: l.TotalLatency,
	}

	return obj
}
