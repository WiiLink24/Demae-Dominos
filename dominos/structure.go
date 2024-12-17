package dominos

import "strconv"

type ToppingGroup int

const (
	Sauce ToppingGroup = iota
	Meat
	NonMeat
	Side
)

func (t *ToppingGroup) New(i string) {
	intValue, _ := strconv.ParseInt(i, 10, 32)
	switch intValue {
	case 0:
		*t = Sauce
	case 1:
		*t = Meat
	case 2:
		*t = NonMeat
	case 3:
		*t = Side
	}
}

type Store struct {
	StoreID      string
	Address      string
	WaitTime     float64
	MinPrice     float64
	IsOpen       bool
	DetailedWait string
	Phone        string
	ServiceHours []ServiceHours
	Information  string
}

type ServiceHours struct {
	OpenTime  string
	CloseTime string
}

type MenuCategory struct {
	Name       string
	Code       string
	Categories []MenuCategory
}

type Item struct {
	Name        string
	Description string
	Img         string
	Items       []ItemSize
}

type ItemSize struct {
	Code  string
	Name  string
	Price string
}

type Topping struct {
	Code  string
	Name  string
	Group ToppingGroup
}

type BasketItem struct {
	Code     string
	Name     string
	Price    float64
	Amount   float64
	Quantity int
	ID       int
	Options  []string
}

type Basket struct {
	Items       []BasketItem
	BasketPrice float64
	ChargePrice float64
	TotalPrice  float64
	OrderId     string
}

type User struct {
	Street          string
	ApartmentNumber string
	City            string
	Region          string
	PostalCode      string
	LocationType    string
	StreetName      string
	StreetNumber    string
	StoreId         string
	Products        []map[string]any
	FirstName       string
	LastName        string
	Email           string
	PhoneNumber     string
	OrderId         string
	Price           string
	OrderTime       string
}
