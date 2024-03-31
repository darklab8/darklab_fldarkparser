package configs_export

import (
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/universe_mapped"
)

type GoodAtBase struct {
	BaseNickname   string
	BaseName       string
	BaseSells      bool
	Price          int
	PricePerVolume float64
	LevelRequired  int
	RepRequired    float64
	SystemName     string
	Faction        string
}

type Commodity struct {
	Nickname            string
	Name                string
	Price               int
	PricePerVolume      float64
	Combinable          bool
	Volume              float64
	NameID              int
	InfocardID          int
	Infocard            InfocardKey
	Bases               []GoodAtBase
	BestBuyPricePerVol  float64
	BestSellPricePerVol float64
	ProffitMarginPerVol float64
}

func (e *Exporter) GetCommodities() []Commodity {
	commodities := make([]Commodity, 0, 100)

	for _, comm := range e.configs.Goods.Commodities {

		var name string
		commodity := Commodity{}
		commodity.Nickname = comm.Nickname.Get()
		commodity.Combinable = comm.Combinable.Get()

		equipment_name := comm.Equipment.Get()
		equipment := e.configs.Equip.CommoditiesMap[equipment_name]

		commodity.NameID = equipment.IdsName.Get()
		if infoname, ok := e.configs.Infocards.Infonames[equipment.IdsName.Get()]; ok {
			name = string(infoname)
		}
		commodity.Name = name
		commodity.Infocard = InfocardKey(commodity.Nickname)
		e.infocards_parser.Set(commodity.Infocard, equipment.IdsInfo.Get())
		commodity.InfocardID = equipment.IdsInfo.Get()

		volume := equipment.Volume.Get()
		commodity.Volume = volume
		commodity.Price = comm.Price.Get()
		if volume != 0 {
			commodity.PricePerVolume = float64(commodity.Price) / float64(volume)
		} else {
			commodity.PricePerVolume = -1
		}

		commodity.Bases = e.GetAtBasesSold(GetAtBasesInput{
			Nickname:       commodity.Nickname,
			Price:          commodity.Price,
			PricePerVolume: commodity.PricePerVolume,
		})

		for _, base_info := range commodity.Bases {
			if base_info.PricePerVolume > float64(commodity.BestSellPricePerVol) {
				commodity.BestSellPricePerVol = base_info.PricePerVolume
			}

			if (base_info.PricePerVolume < commodity.BestBuyPricePerVol || commodity.BestBuyPricePerVol == 0) && base_info.BaseSells {
				commodity.BestBuyPricePerVol = base_info.PricePerVolume
			}
		}

		if commodity.BestBuyPricePerVol > 0 && commodity.BestSellPricePerVol > 0 {
			commodity.ProffitMarginPerVol = commodity.BestSellPricePerVol - commodity.BestBuyPricePerVol
		}

		commodities = append(commodities, commodity)
	}

	return commodities
}

type GetAtBasesInput struct {
	Nickname       string
	Price          int
	PricePerVolume float64
}

func (e *Exporter) GetAtBasesSold(commodity GetAtBasesInput) []GoodAtBase {
	var bases_list []GoodAtBase

	for _, base_market := range e.configs.Market.BasesPerGood[commodity.Nickname] {
		base_nickname := base_market.Base
		market_good := base_market.MarketGood
		base_info := GoodAtBase{}
		base_info.BaseSells = !market_good.IsBuyOnly.Get()
		base_info.BaseNickname = base_nickname
		base_info.Price = int(market_good.PriceModifier.Get() * float64(commodity.Price))
		base_info.PricePerVolume = market_good.PriceModifier.Get() * float64(commodity.PricePerVolume)

		base_info.LevelRequired = market_good.LevelRequired.Get()
		base_info.RepRequired = market_good.RepRequired.Get()

		if universe_base, ok := e.configs.Universe_config.BasesMap[universe_mapped.BaseNickname(base_nickname)]; ok {

			if infoname, ok := e.configs.Infocards.Infonames[universe_base.StridName.Get()]; ok {
				base_info.BaseName = string(infoname)
			}
			system_nickname := universe_base.System.Get()

			if system, ok := e.configs.Universe_config.SystemMap[universe_mapped.SystemNickname(system_nickname)]; ok {
				if infoname, ok := e.configs.Infocards.Infonames[system.Strid_name.Get()]; ok {
					base_info.SystemName = string(infoname)
				}
			}

			var reputation_nickname string
			if system, ok := e.configs.Systems.SystemsMap[universe_base.System.Get()]; ok {
				for _, system_base := range system.Bases {
					if system_base.IdsName.Get() == universe_base.StridName.Get() {
						reputation_nickname = system_base.RepNickname.Get()
					}
				}
			}

			var factionName string
			if group, exists := e.configs.InitialWorld.GroupsMap[reputation_nickname]; exists {
				if faction_name, exists := e.configs.Infocards.Infonames[group.IdsName.Get()]; exists {
					factionName = string(faction_name)
				}
			}

			base_info.Faction = factionName

			bases_list = append(bases_list, base_info)
		}
	}
	return bases_list
}
