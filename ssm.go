package main

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type SSMParameter struct {
	Name    string
	Value   string
	Version int64
}

func (p SSMParameter) String() string {
	return p.Value
}

func SSMFromParameter(p *ssm.Parameter) SSMParameter {
	return SSMParameter{
		Name:    *p.Name,
		Value:   *p.Value,
		Version: *p.Version,
	}
}

var _ssmSvc *ssm.SSM = nil

func getSvc() *ssm.SSM {
	if _ssmSvc == nil {
		sess := session.Must(session.NewSession())
		_ssmSvc = ssm.New(sess)
	}

	return _ssmSvc
}

func ssmGet(name string, params ParamValue) (*SSMParameter, error) {
	var err error
	input := ssm.GetParameterInput{Name: &name}

	if b, ok := params["decrypt"]; ok {
		input.SetWithDecryption(b.Bool())
	}

	if err = input.Validate(); err != nil {
		return nil, err
	}

	p, err := getSvc().GetParameter(&input)

	if err != nil {
		return nil, err
	}

	ret := SSMFromParameter(p.Parameter)
	return &ret, nil
}

func ssmGetPath(path string, params ParamValue) ([]SSMParameter, error) {
	var err error
	var trim bool = false

	input := ssm.GetParametersByPathInput{Path: &path}

	if b, ok := params["decrypt"]; ok {
		input.SetWithDecryption(b.Bool())
	}

	if b, ok := params["recurse"]; ok {
		input.SetRecursive(b.Bool())
	}

	if i, ok := params["maxresults"]; ok {
		input.SetMaxResults(i.Int())
	}

	if b, ok := params["trim"]; ok {
		trim = b.Bool()
	}

	// TODO: array of string filters

	if err = input.Validate(); err != nil {
		return nil, err
	}

	var output []SSMParameter

	err = getSvc().GetParametersByPathPages(&input,
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, p := range page.Parameters {
				ssp := SSMFromParameter(p)
				if trim {
					ssp.Name = strings.TrimPrefix(*p.Name, path)
				}
				output = append(output, ssp)
			}

			return true
		})

	return output, err
}

func getSSMFuncMap() template.FuncMap {
	return template.FuncMap{
		"ssmGet": func(path string, args ...string) (*SSMParameter, error) {
			tmpl, err := ParamParse(ParamDecl{
				"decrypt": reflect.Bool,
			}, args)

			if err != nil {
				return nil, err
			}

			return ssmGet(path, tmpl)
		},

		"ssmGetPath": func(path string, args ...string) ([]SSMParameter, error) {
			tmpl, err := ParamParse(ParamDecl{
				"decrypt":    reflect.Bool,
				"maxresults": reflect.Int,
				"recursive":  reflect.Bool,
				"trim":       reflect.Bool,
			}, args)

			if err != nil {
				return nil, err
			}

			return ssmGetPath(path, tmpl)
		},
	}
}
