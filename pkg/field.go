package pkg


//struct field
type Field struct {
	Name string

	Type string

	TypeName string

	//Name + type or type
	Field string

	Size int

	Doc []string

	Tag string
}

type Fields []Field


func (f Fields) Len() int { return len(f) }

func (f Fields) Less(i, j int) bool {
	return f[i].Size <  f[j].Size
}

func (f Fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
