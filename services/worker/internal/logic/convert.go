package logic

import (
	"worker/internal/store"
	"worker/worker"
)

func userToPb(u store.User) *worker.User {
	return userToPbWithBuildings(u, nil)
}

func userToPbWithBuildings(u store.User, buildingIDs []int64) *worker.User {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	buildingID := int64(0)
	if u.BuildingID.Valid {
		buildingID = u.BuildingID.Int64
	}
	maxConcurrent := u.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	return &worker.User{
		Id:            u.ID,
		Username:      u.Username,
		Role:          u.Role,
		Name:          u.Name,
		Phone:         phone,
		BuildingId:    buildingID,
		Status:        u.Status,
		BuildingIds:   buildingIDs,
		MaxConcurrent: maxConcurrent,
	}
}

func workerInfoToPb(u store.User, baseBuildingID int64, skills []store.Skill) *worker.WorkerInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	info := &worker.WorkerInfo{
		Id:             u.ID,
		Username:       u.Username,
		Name:           u.Name,
		Phone:          phone,
		BaseBuildingId: baseBuildingID,
	}
	if u.MaxConcurrent > 0 {
		info.MaxConcurrent = u.MaxConcurrent
	} else {
		info.MaxConcurrent = 3
	}
	for _, s := range skills {
		info.Skills = append(info.Skills, &worker.SkillInfo{
			Name:        s.Name,
			Proficiency: s.Proficiency,
		})
	}
	return info
}
