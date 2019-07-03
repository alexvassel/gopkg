package job

import (
	"context"
	"reflect"

	"github.com/gocraft/work"
	"github.com/severgroup-tt/gopkg-errors"
)

const nameProperty = "jobName"

type WorkRecord struct {
	Job      interface{}
	Fn       interface{}
	Schedule string
}

func registerPeriodicalJob(ctx context.Context, pool *work.WorkerPool, job interface{}, fn interface{}, schedule string) error {
	if err := registerJob(ctx, pool, job, fn); err != nil {
		return err
	}
	jobName, _ := getJobName(ctx, job)
	pool.PeriodicallyEnqueue(schedule, jobName)
	return nil
}

func registerJob(ctx context.Context, pool *work.WorkerPool, job interface{}, fn interface{}) error {
	jobName, err := getJobName(ctx, job)
	if err != nil {
		return err
	}

	pool.Job(jobName, fn)
	return nil
}

func getJobName(ctx context.Context, job interface{}) (string, error) {
	st := reflect.TypeOf(job)
	field, ok := st.FieldByName(nameProperty)
	if !ok {
		return "", errors.Internal.Err(ctx, "Not fount property '"+nameProperty+"' in Job container")
	}
	jobName := field.Tag.Get("job")
	if jobName == "" {
		return "", errors.Internal.Err(ctx, "Empty property '"+nameProperty+"' in Job container")
	}
	return jobName, nil
}

func buildArgs(ctx context.Context, job interface{}) (map[string]interface{}, error) {
	st := reflect.ValueOf(job)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	numFields := st.NumField()
	ret := make(map[string]interface{}, numFields)
	for i := 0; i < numFields; i++ {
		name := st.Type().Field(i).Name
		if name == nameProperty {
			continue
		}
		kind := st.Type().Field(i).Type.Kind()
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String, reflect.Bool:
			ret[st.Type().Field(i).Name] = st.Field(i).Interface()
		default:
			return nil, errors.Internal.Err(ctx, "Unsupported argument type").
				WithLogKV("arg", name, "kind", kind)
		}

	}
	return ret, nil
}

func fillArgs(ctx context.Context, job interface{}, workJob *work.Job) error {
	st := reflect.ValueOf(job)
	if st.Kind() != reflect.Ptr {
		return errors.Internal.Err(ctx, "Job should be pointer")
	}
	st = st.Elem()
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		name := st.Type().Field(i).Name
		if name == nameProperty {
			continue
		}
		value, ok := workJob.Args[name]
		if !ok {
			continue
		}
		var castOk bool
		kind := st.Type().Field(i).Type.Kind()
		switch kind {
		case reflect.Int:
			if v, ok := value.(int); ok {
				st.Field(i).SetInt(int64(v))
				castOk = true
			}
		case reflect.Int8:
			if v, ok := value.(int8); ok {
				st.Field(i).SetInt(int64(v))
				castOk = true
			}
		case reflect.Int16:
			if v, ok := value.(int16); ok {
				st.Field(i).SetInt(int64(v))
				castOk = true
			}
		case reflect.Int32:
			if v, ok := value.(int32); ok {
				st.Field(i).SetInt(int64(v))
				castOk = true
			}
		case reflect.Int64:
			if v, ok := value.(int64); ok {
				st.Field(i).SetInt(v)
				castOk = true
			}
		case reflect.String:
			if v, ok := value.(string); ok {
				st.Field(i).SetString(v)
				castOk = true
			}
		case reflect.Bool:
			if v, ok := value.(bool); ok {
				st.Field(i).SetBool(v)
				castOk = true
			}
		default:
			return errors.Internal.Err(ctx, "Cant cast Job argument: not implemented").
				WithLogKV("name", name, "kind", kind, "value", workJob.Args[name])
		}
		if !castOk {
			return errors.Internal.Err(ctx, "Cant cast Job argument").
				WithLogKV("name", name, "kind", kind, "value", workJob.Args[name])
		}
	}
	return nil
}
