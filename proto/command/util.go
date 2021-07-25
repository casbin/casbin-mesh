package command

import "encoding/json"

func NewStringArray(input [][]string) []*StringArray {
	var out []*StringArray
	for _, s := range input {
		out = append(out, &StringArray{
			S: s,
		})
	}
	return out
}

func ToStringArray(input []*StringArray) [][]string {
	var out [][]string
	for _, i := range input {
		out = append(out, i.GetS())
	}
	return out
}

func ToInterfaces(input [][]byte) []interface{} {
	var params []interface{}
	var err error
	for _, b := range input {
		var tmp interface{}
		err = json.Unmarshal(b, &tmp)
		if err != nil {
			return nil
		}
		params = append(params, tmp)
	}
	return params
}
