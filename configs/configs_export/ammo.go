package configs_export

import "github.com/darklab8/go-utils/utils/ptr"

type Ammo struct {
	Name  string
	Price int

	HitPts           int
	Volume           float64
	MunitionLifetime float64

	Nickname     string
	NameID       int
	InfoID       int
	SeekerType   string
	SeekerRange  int
	SeekerFovDeg int

	Bases []*GoodAtBase

	*DiscoveryTechCompat

	AmmoLimit AmmoLimit
}

func (e *Exporter) GetAmmo(ids []Tractor) []Ammo {
	var tractors []Ammo

	for _, munition_info := range e.configs.Equip.Munitions {
		munition := Ammo{}
		munition.Nickname = munition_info.Nickname.Get()
		munition.NameID, _ = munition_info.IdsName.GetValue()
		munition.InfoID, _ = munition_info.IdsInfo.GetValue()

		munition.HitPts, _ = munition_info.HitPts.GetValue()

		if value, ok := munition_info.AmmoLimitAmountInCatridge.GetValue(); ok {
			munition.AmmoLimit.AmountInCatridge = ptr.Ptr(value)
		}
		if value, ok := munition_info.AmmoLimitMaxCatridges.GetValue(); ok {
			munition.AmmoLimit.MaxCatridges = ptr.Ptr(value)
		}

		munition.Volume, _ = munition_info.Volume.GetValue()
		munition.SeekerRange, _ = munition_info.SeekerRange.GetValue()
		munition.SeekerType, _ = munition_info.SeekerType.GetValue()

		munition.MunitionLifetime, _ = munition_info.LifeTime.GetValue()

		munition.SeekerFovDeg, _ = munition_info.SeekerFovDeg.GetValue()

		if ammo_ids_name, ok := munition_info.IdsName.GetValue(); ok {
			munition.Name = e.GetInfocardName(ammo_ids_name, munition.Nickname)
		}

		munition.Price = -1
		if good_info, ok := e.configs.Goods.GoodsMap[munition_info.Nickname.Get()]; ok {
			if price, ok := good_info.Price.GetValue(); ok {
				munition.Price = price
				munition.Bases = e.GetAtBasesSold(GetAtBasesInput{
					Nickname: good_info.Nickname.Get(),
					Price:    price,
				})
			}
		}

		if !e.Buyable(munition.Bases) && (munition.Name == "") {
			continue
		}

		e.exportInfocards(InfocardKey(munition.Nickname), munition.InfoID)
		munition.DiscoveryTechCompat = CalculateTechCompat(e.configs.Discovery, ids, munition.Nickname)
		tractors = append(tractors, munition)
	}
	return tractors
}

func (e *Exporter) FilterToUsefulAmmo(cms []Ammo) []Ammo {
	var useful_items []Ammo = make([]Ammo, 0, len(cms))
	for _, item := range cms {
		if !e.Buyable(item.Bases) {
			continue
		}
		useful_items = append(useful_items, item)
	}
	return useful_items
}
