package yamlenv

import (
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

var reVar = regexp.MustCompile(`\$\{([^{}]+)}`)

type Env[T int | string | bool] struct {
	Value T
}

func (e *Env[T]) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	var ret T
	var val string
	if match := reVar.FindString(s); match != "" {
		val = string(e.replaceEnvString([]byte(s)))
	} else {
		val = s
	}
	switch p := any(&ret).(type) {
	case *string:
		*p = val
		e.Value = ret
	case *int:
		i, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*p = i
		e.Value = ret
	case *bool:
		i, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*p = i
		e.Value = ret
	}
	return nil
}

func (e *Env[T]) replaceEnvString(val []byte) []byte {
	return reVar.ReplaceAllFunc(val, func(b []byte) []byte {
		group1 := reVar.ReplaceAllString(string(b), `$1`)

		envValue := os.Getenv(group1)
		if len(envValue) > 0 {
			return []byte(envValue)
		}
		return []byte("")
	})
}
