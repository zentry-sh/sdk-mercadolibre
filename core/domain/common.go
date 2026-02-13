package domain

import "time"

type Money struct {
	Amount   float64
	Currency string
}

func NewMoney(amount float64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) IsPositive() bool {
	return m.Amount > 0
}

type Address struct {
	Street    string
	Number    string
	Floor     string
	Apartment string
	City      string
	State     string
	ZipCode   string
	Country   string
	Lat       float64
	Lon       float64
}

func (a Address) IsEmpty() bool {
	return a.Street == "" && a.City == "" && a.Country == ""
}

type Payer struct {
	ID             string
	Email          string
	FirstName      string
	LastName       string
	Phone          string
	Identification Identification
	Address        *Address
}

func (p Payer) FullName() string {
	if p.FirstName == "" && p.LastName == "" {
		return ""
	}
	if p.LastName == "" {
		return p.FirstName
	}
	if p.FirstName == "" {
		return p.LastName
	}
	return p.FirstName + " " + p.LastName
}

type Identification struct {
	Type   string
	Number string
}

func (i Identification) IsEmpty() bool {
	return i.Type == "" && i.Number == ""
}

type Package struct {
	Weight      float64
	Length      float64
	Width       float64
	Height      float64
	Description string
}

func (p Package) VolumetricWeight() float64 {
	return (p.Length * p.Width * p.Height) / 5000
}

type Carrier struct {
	ID          string
	Name        string
	ServiceType string
}

type LabelInfo struct {
	URL       string
	Format    string
	CreatedAt time.Time
}

type Dimensions struct {
	Length float64
	Width  float64
	Height float64
}
