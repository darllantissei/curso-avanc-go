package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Environment int64

const (
	environment_Undefined Environment = iota - 1
	environment_Sandbox
	environment_Production

	environment_name_undefined  = "undefined"
	environment_name_sandbox    = "sandbox"
	environment_name_production = "production"

	environment_text_enus_undefined  = "The environment is undefined"
	environment_text_enus_sandbox    = "The environment is sandbox"
	environment_text_enus_production = "The environment is production"

	environment_text_ptbr_undefined  = "Ambiente indefinido"
	environment_text_ptbr_sandbox    = "Ambiente de sandbox"
	environment_text_ptbr_production = "Ambiente de produção"
)

var (
	environment_name = map[int64]string{
		-1: environment_name_undefined,
		0:  environment_name_sandbox,
		1:  environment_name_production,
	}
	environment_value = map[string]int64{
		environment_name_undefined:  -1,
		environment_name_sandbox:    0,
		environment_name_production: 1,
	}
)

func (e Environment) String() string {

	switch e {
	case environment_Undefined:
		return environment_name_undefined
	case environment_Sandbox:
		return environment_name_sandbox
	case environment_Production:
		return environment_name_production
	default:
		panic("type environment is invalid")
	}
}

func (e Environment) MessageEN() string {

	switch e {
	case environment_Undefined:
		return environment_text_enus_undefined
	case environment_Sandbox:
		return environment_text_enus_sandbox
	case environment_Production:
		return environment_text_enus_production
	default:
		panic("type environment is invalid")
	}
}

func (e Environment) MessagePT() string {

	switch e {
	case environment_Undefined:
		return environment_text_ptbr_undefined
	case environment_Sandbox:
		return environment_text_ptbr_sandbox
	case environment_Production:
		return environment_text_ptbr_production
	default:
		panic("type environment is invalid")
	}
}

func (e Environment) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.String())), nil
}

func (e *Environment) UnmarshalJSON(bytes []byte) error {
	value, err := e.tryGetValueFromJSON(bytes)
	if err == nil && !strings.EqualFold(value, "") {

		env, err := e.tryParseValueToEnvironment(value)

		if err != nil {
			return err
		}

		*e = Environment(env)
	}

	return err
}
func (e *Environment) Scan(value interface{}) error {

	switch data := value.(type) {
	case []uint8:
		str := string([]byte(data))

		env, err := e.tryParseValueToEnvironment(str)

		if err != nil {
			return err
		}

		*e = Environment(env)

	case int:
		d := int64(data)
		*e = Environment(d)
	case int32:
		d := int64(data)
		*e = Environment(d)
	case float32:
		d := int64(data)
		*e = Environment(d)
	case float64:
		d := int64(data)
		*e = Environment(d)
	case int64:
		*e = Environment(data)
	case string:
		env, err := e.tryParseValueToEnvironment(data)
		if err != nil {
			panic(err)
		}
		*e = env
	}

	return nil
}

func (e *Environment) tryParseValueToEnvironment(value string) (env Environment, err error) {

	envINT, ok := environment_value[value]

	if !ok {
		valueINT, err := strconv.Atoi(value)
		if err != nil {
			return -1, fmt.Errorf("the %s is incorret to type of environment", value)
		}

		envStr, ok := environment_name[int64(valueINT)]
		if !ok {
			return -1, fmt.Errorf("the %s not valid type of environment", value)
		}

		envINT, ok = environment_value[envStr]
		if !ok {
			return -1, fmt.Errorf("the %s is invalid type of environment", value)
		}
	}

	return Environment(envINT), nil
}

func (e *Environment) tryGetValueFromJSON(bytes []byte) (value string, err error) {
	value, err = strconv.Unquote(string(bytes))

	if err != nil {

		valueINT := int(-1)

		valueINT, err = strconv.Atoi(string(bytes))

		if err != nil {
			return
		}

		value = fmt.Sprintf("%d", valueINT)
		err = nil
	}

	return
}
