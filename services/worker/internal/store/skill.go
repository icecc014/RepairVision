package store

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Skill struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Proficiency int64  `db:"proficiency"`
}

func EnsureSkill(ctx context.Context, conn sqlx.Session, name string) (int64, error) {
	if _, err := conn.ExecCtx(ctx,
		"insert ignore into skills(name) values(?)", name); err != nil {
		return 0, err
	}
	var id int64
	if err := conn.QueryRowCtx(ctx, &id,
		"select id from skills where name = ?", name); err != nil {
		return 0, err
	}
	return id, nil
}

func EnsureWorkerSkill(ctx context.Context, conn sqlx.Session, workerID, skillID, proficiency int64) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into worker_skills(worker_id, skill_id, proficiency) values(?,?,?)",
		workerID, skillID, proficiency)
	return err
}

func ListWorkerSkills(ctx context.Context, conn sqlx.Session, workerID int64) ([]Skill, error) {
	var skills []Skill
	if err := conn.QueryRowsCtx(ctx, &skills,
		`select s.id, s.name, ws.proficiency
		 from worker_skills ws join skills s on s.id = ws.skill_id
		 where ws.worker_id = ? order by s.id`, workerID); err != nil {
		return nil, err
	}
	return skills, nil
}

func ErrNoRows() error {
	return errors.New("no rows")
}

var _ = ErrNoRows
