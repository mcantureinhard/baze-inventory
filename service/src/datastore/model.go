package datastore

type Micronutrient struct {
	BaseUnit       string  `json:"base_unit,omitempty"`
	Name           string  `json:"name,omitempty"`
	Quantity       float64 `json:"quantity,omitempty"`
	UnitMultiplier float64 `json:"unit_multiplier,omitempty"`
}

type PillMicronutrient struct {
	MicroNutrient *Micronutrient `json:"micronutrient"`
	Absortion     int            `json:"absortion_percent"`
}

type Pill struct {
	Name               string               `json:"name"`
	PillMicronutrients []*PillMicronutrient `json:"pillmicronutrients,omitempty"`
}

type Pills []Pill

type Micronutrients []Micronutrient

type PillInventory struct {
	PillData  *Pill `json:pill`
	Inventory int   `json:"inventory,omitempty"`
}

type PillInventories []PillInventory
