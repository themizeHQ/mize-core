package types

import (
	"strconv"
	"strings"
	"time"

	"mize.app/app/schedule/models"
)

type SchedulePayload struct {
	Payload models.Schedule `bson:"payload" json:"payload"`
	Date    string          `bson:"date" json:"date"`
	Time    string          `bson:"time" json:"time"`
}

func (sp *SchedulePayload) SetScheduleTime() error {
	parsedDate, err := time.Parse("2006-01-02", sp.Date)
	if err != nil {
		return err
	}
	parsedTime := strings.Split(sp.Time, ":")
	hr, err := strconv.Atoi(parsedTime[0])
	if err != nil {
		return err
	}
	min, err := strconv.Atoi(parsedTime[1])
	if err != nil {
		return err
	}
	sp.Payload.Time = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Second(), hr, min, 0, 0, parsedDate.Location()).Unix()
	return nil
}
