package workout

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Job struct {
	Id        uint64        `json:"id"`
	Priority  uint32        `json:"priority"`
	Tube      string        `json:"tube"`
	TimeToRun time.Duration `json:"ttr"`
	Body      string        `json:"body"`
	Age       time.Duration `json:"age"`
	Attempt   uint32        `json:"failure_count"`
}

func (j *Job) NextDelay() time.Duration {
	d1 := int(math.Pow(float64(j.Attempt), 4.0))
	d2 := rand.Intn(30) * int(j.Attempt)
	return time.Duration(d1+d2+15) * time.Second
}

func (j *Job) Describe() string {
	return fmt.Sprintf("%d (tube=%s pri=%d ttr=%v age=%v attempt=%d)", j.Id, j.Tube, j.Priority, j.TimeToRun, j.Age, j.Attempt)
}
