package command

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
