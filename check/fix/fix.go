package fix

type Diff interface {
	NameStr() string

	CurrentStr() string
	DesiredStr() string

	HasCurrent() bool
}

type SimpleDiff struct {
	Name string

	Current string
	Desired string

	MissingCurrent bool
}

func (d SimpleDiff) NameStr() string { return d.Name }

func (d SimpleDiff) CurrentStr() string { return d.Current }
func (d SimpleDiff) DesiredStr() string { return d.Desired }

func (d SimpleDiff) HasCurrent() bool { return !d.MissingCurrent }

type Fix interface {
	Diff
	Fix() error
}
