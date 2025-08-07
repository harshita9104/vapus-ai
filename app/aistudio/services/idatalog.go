package services

import "context"

type IDataLog interface {
	Create(context.Context) error
	GetById()
	GetAll()
}
