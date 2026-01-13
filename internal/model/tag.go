package model

type Tag struct {
	ID    int64
	Name  string
	Color string
}

func NewTag(name string) *Tag {
	return &Tag{
		Name:  name,
		Color: "#808080",
	}
}

var DefaultTagColors = []string{
	"#E57373", // red
	"#81C784", // green
	"#64B5F6", // blue
	"#FFD54F", // yellow
	"#BA68C8", // purple
	"#4DD0E1", // cyan
	"#FF8A65", // orange
	"#A1887F", // brown
}
