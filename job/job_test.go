package job

import (
	"context"
	"testing"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
)

type testJob struct {
	jobName struct{} `Job:"job_name"`
	Int     int
	Int32   int32
	Int64   int64
	String  string
	Bool    bool
}

func TestJob(t *testing.T) {
	job := testJob{
		Int:    1,
		Int32:  2,
		Int64:  3,
		String: "4",
		Bool:   true,
	}
	args := map[string]interface{}{
		"Int":    int(1),
		"Int32":  int32(2),
		"Int64":  int64(3),
		"String": "4",
		"Bool":   true,
	}

	t.Run("Build args", func(t *testing.T) {
		gotArgs, err := buildArgs(context.Background(), job)

		assert.Nil(t, err)
		assert.Equal(t, args, gotArgs)
	})

	t.Run("Build args pointer", func(t *testing.T) {
		gotArgs, err := buildArgs(context.Background(), &job)
		assert.Nil(t, err)
		assert.Equal(t, args, gotArgs)
	})

	t.Run("Fill args", func(t *testing.T) {
		gotJob := testJob{}
		err := fillArgs(context.Background(), &gotJob, &work.Job{Args: args})
		assert.Nil(t, err)
		assert.Equal(t, job, gotJob)
	})
}
