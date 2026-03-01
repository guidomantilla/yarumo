package products

var _ Gun = (*DefaultGun)(nil)

//

type Gun interface {
	SetName(name string)
	SetPower(power int)
	GetName() string
	GetPower() int
}

type DefaultGun struct {
	name  string
	power int
}

func (g *DefaultGun) SetName(name string) {
	g.name = name
}

func (g *DefaultGun) GetName() string {
	return g.name
}

func (g *DefaultGun) SetPower(power int) {
	g.power = power
}

func (g *DefaultGun) GetPower() int {
	return g.power
}

//

type Ak47 struct {
	DefaultGun
}

func NewAk47() Gun {
	return &Ak47{
		DefaultGun: DefaultGun{
			name:  "AK47 gun",
			power: 4,
		},
	}
}

//

type Musket struct {
	DefaultGun
}

func NewMusket() Gun {
	return &Musket{
		DefaultGun: DefaultGun{
			name:  "Musket gun",
			power: 1,
		},
	}
}
