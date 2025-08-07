package models

import (
	mpb "github.com/vapusdata-ecosystem/apis/protos/models/v1alpha1"
)

type CacheLogs struct {
	ID           int64  `json:"id" bun:"id,pk,autoincrement"`
	RequestedAt  int64  `json:"requestedAt" bun:"requested_at,notnull"`
	Endpoint     string `json:"endpoint,omitempty" bun:"endpoint"`
	Operation    string `json:"operation" bun:"operation,notnull"`
	Status       string `json:"status,omitempty" bun:"status"`
	Key          string `json:"key" bun:"key,notnull"`
	TotalLatency int64  `json:"totalLatency" bun:"total_latency,notnull"`
}

func (l *CacheLogs) ConvertFromPb(pb *mpb.CacheLog) *CacheLogs {
	if l == nil {
		return nil
	}

	obj := &CacheLogs{
		RequestedAt:  pb.RequestedAt,
		Endpoint:     pb.Endpoint,
		Operation:    pb.Operation.String(),
		Status:       pb.Status.String(),
		Key:          pb.Key,
		TotalLatency: pb.TotalLatency,
	}

	return obj
}

func (l *CacheLogs) ConvertToPb() *mpb.CacheLog {
	if l == nil {
		return nil
	}

	obj := &mpb.CacheLog{
		RequestedAt:  l.RequestedAt,
		Endpoint:     l.Endpoint,
		Operation:    mpb.CacheOperationType(mpb.CacheOperationType_value[l.Operation]),
		Status:       mpb.CacheStatusType(mpb.CacheOperationType_value[l.Status]),
		Key:          l.Key,
		TotalLatency: l.TotalLatency,
	}

	return obj
}
