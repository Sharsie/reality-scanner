package reality

type KnownReality struct {
	Deletable bool
	Id        string
	Images    []string
	IsNew     bool
	Link      string
	Place     string
	Price     int
	Title     string
}

type KnownRealities map[string]KnownReality
