package opencast

type Field struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

type Metadata struct {
	Flavor string  `json:"flavor"`
	Fields []Field `json:"fields"`
}
