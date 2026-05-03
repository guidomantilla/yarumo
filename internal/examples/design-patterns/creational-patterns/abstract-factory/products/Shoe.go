package products

var _ Shoe = (*AbstractShoe)(nil)

var _ Shoe = (*AdidasShoe)(nil)

var _ Shoe = (*NikeShoe)(nil)

//

type Shoe interface {
	SetLogo(logo string)
	GetLogo() string

	SetSize(size int)
	GetSize() int
}

type AbstractShoe struct {
	Logo string
	Size int
}

func (s *AbstractShoe) SetLogo(logo string) {
	s.Logo = logo
}

func (s *AbstractShoe) GetLogo() string {
	return s.Logo
}

func (s *AbstractShoe) SetSize(size int) {
	s.Size = size
}

func (s *AbstractShoe) GetSize() int {
	return s.Size
}

//

type AdidasShoe struct {
	Shoe
}

type NikeShoe struct {
	Shoe
}
