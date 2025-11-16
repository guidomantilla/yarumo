package products

var _ Shirt = (*AbstractShirt)(nil)

var _ Shirt = (*AdidasShirt)(nil)

var _ Shirt = (*NikeShirt)(nil)

//

type Shirt interface {
	SetLogo(logo string)
	GetLogo() string

	SetSize(size int)
	GetSize() int
}

//

type AbstractShirt struct {
	Logo string
	Size int
}

func (s *AbstractShirt) SetLogo(logo string) {
	s.Logo = logo
}

func (s *AbstractShirt) GetLogo() string {
	return s.Logo
}

func (s *AbstractShirt) SetSize(size int) {
	s.Size = size
}

func (s *AbstractShirt) GetSize() int {
	return s.Size
}

//

type AdidasShirt struct {
	Shirt
}

type NikeShirt struct {
	Shirt
}
