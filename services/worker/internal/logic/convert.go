package logic

import (
	"worker/internal/store"
	"worker/worker"
)

func userToPb(u store.User) *worker.User {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	buildingID := int64(0)
	if u.BuildingID.Valid {
		buildingID = u.BuildingID.Int64
	}
	return &worker.User{
		Id:         u.ID,
		Username:   u.Username,
		Role:       u.Role,
		Name:       u.Name,
		Phone:      phone,
		BuildingId: buildingID,
		Status:     u.Status,
	}
}

func workerInfoToPb(u store.User) *worker.WorkerInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	return &worker.WorkerInfo{
		Id:       u.ID,
		Username: u.Username,
		Name:     u.Name,
		Phone:    phone,
	}
}
