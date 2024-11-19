package configs_export

import (
	"fmt"
	"strings"

	"github.com/darklab8/fl-configs/configs/cfgtype"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/initialworld/flhash"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/universe_mapped"
)

type GoodAtBase struct {
	BaseNickname      cfgtype.BaseUniNick
	BaseSells         bool
	PriceBaseBuysFor  int
	PriceBaseSellsFor int
	Volume            float64
	ShipClass         cfgtype.ShipClass
	LevelRequired     int
	RepRequired       float64

	NotBuyable           bool
	IsServerSideOverride bool

	IsTransportUnreachable bool

	BaseInfo
}

type Commodity struct {
	Nickname              string
	NicknameHash          flhash.HashCode
	Name                  string
	Combinable            bool
	Volume                float64
	ShipClass             cfgtype.ShipClass
	NameID                int
	InfocardID            int
	Infocard              InfocardKey
	Bases                 map[cfgtype.BaseUniNick]*GoodAtBase
	PriceBestBaseBuysFor  int
	PriceBestBaseSellsFor int
	ProffitMargin         int
	BaseAllTradeRoutes
}

func GetPricePerVoume(price int, volume float64) float64 {
	if volume == 0 {
		return -1
	}
	return float64(price) / float64(volume)
}

func (e *Exporter) GetCommodities() []*Commodity {
	commodities := make([]*Commodity, 0, 100)

	for _, comm := range e.configs.Goods.Commodities {
		equipment_name := comm.Equipment.Get()
		equipment := e.configs.Equip.CommoditiesMap[equipment_name]

		for _, volume_info := range equipment.Volumes {
			commodity := &Commodity{
				Bases: make(map[cfgtype.BaseUniNick]*GoodAtBase),
			}
			commodity.Nickname = comm.Nickname.Get()
			commodity.NicknameHash = flhash.HashNickname(commodity.Nickname)
			e.Hashes[commodity.Nickname] = commodity.NicknameHash

			commodity.Combinable = comm.Combinable.Get()

			commodity.NameID = equipment.IdsName.Get()

			commodity.Name = e.GetInfocardName(equipment.IdsName.Get(), commodity.Nickname)
			e.exportInfocards(commodity.Infocard, equipment.IdsInfo.Get())
			commodity.InfocardID = equipment.IdsInfo.Get()

			commodity.Volume = volume_info.Volume.Get()
			commodity.ShipClass = volume_info.GetShipClass()
			commodity.Infocard = InfocardKey(commodity.Nickname)

			base_item_price := comm.Price.Get()

			commodity.Bases = e.GetAtBasesSold(GetCommodityAtBasesInput{
				Nickname:  commodity.Nickname,
				Price:     base_item_price,
				Volume:    commodity.Volume,
				ShipClass: commodity.ShipClass,
			})

			for _, base_info := range commodity.Bases {
				if base_info.PriceBaseBuysFor > commodity.PriceBestBaseBuysFor {
					commodity.PriceBestBaseBuysFor = base_info.PriceBaseBuysFor
				}
				if base_info.PriceBaseSellsFor < commodity.PriceBestBaseSellsFor || commodity.PriceBestBaseSellsFor == 0 {
					if base_info.BaseSells {
						commodity.PriceBestBaseSellsFor = base_info.PriceBaseSellsFor
					}

				}
			}

			if commodity.PriceBestBaseBuysFor > 0 && commodity.PriceBestBaseSellsFor > 0 {
				commodity.ProffitMargin = commodity.PriceBestBaseBuysFor - commodity.PriceBestBaseSellsFor
			}

			commodities = append(commodities, commodity)
		}

	}

	return commodities
}

type GetCommodityAtBasesInput struct {
	Nickname  string
	Price     int
	Volume    float64
	ShipClass cfgtype.ShipClass
}

func (e *Exporter) ServerSideMarketGoodsOverrides(commodity GetCommodityAtBasesInput) map[cfgtype.BaseUniNick]*GoodAtBase {
	var bases_already_found map[cfgtype.BaseUniNick]*GoodAtBase = make(map[cfgtype.BaseUniNick]*GoodAtBase)

	for _, base_market := range e.configs.Discovery.Prices.BasesPerGood[commodity.Nickname] {
		var base_info *GoodAtBase
		base_nickname := cfgtype.BaseUniNick(base_market.BaseNickname.Get())

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				fmt.Println("recovered base_nickname", base_nickname)
				fmt.Println("recovered commodity nickname", commodity.Nickname)
				panic(r)
			}
		}()

		base_info = &GoodAtBase{
			NotBuyable:           false,
			BaseNickname:         base_nickname,
			BaseSells:            base_market.BaseSells.Get(),
			PriceBaseBuysFor:     base_market.PriceBaseBuysFor.Get(),
			PriceBaseSellsFor:    base_market.PriceBaseSellsFor.Get(),
			Volume:               commodity.Volume,
			ShipClass:            commodity.ShipClass,
			IsServerSideOverride: true,
		}

		base_info.BaseInfo = e.GetBaseInfo(universe_mapped.BaseNickname(base_info.BaseNickname))

		if e.useful_bases_by_nick != nil {
			if _, ok := e.useful_bases_by_nick[base_info.BaseNickname]; !ok {
				base_info.NotBuyable = true
			}
		}

		bases_already_found[base_info.BaseNickname] = base_info
	}
	return bases_already_found
}

func (e *Exporter) GetAtBasesSold(commodity GetCommodityAtBasesInput) map[cfgtype.BaseUniNick]*GoodAtBase {
	var goods_per_base map[cfgtype.BaseUniNick]*GoodAtBase = make(map[cfgtype.BaseUniNick]*GoodAtBase)

	for _, base_market := range e.configs.Market.BasesPerGood[commodity.Nickname] {
		base_nickname := base_market.Base

		market_good := base_market.MarketGood
		base_info := &GoodAtBase{
			NotBuyable: false,
			Volume:     commodity.Volume,
			ShipClass:  commodity.ShipClass,
		}
		base_info.BaseSells = market_good.BaseSells()
		base_info.BaseNickname = base_nickname

		base_info.PriceBaseSellsFor = int(market_good.PriceModifier.Get() * float64(commodity.Price))

		if e.configs.Discovery != nil {
			base_info.PriceBaseBuysFor = market_good.BaseSellsIPositiveAndDiscoSellPrice.Get()
		} else {
			base_info.PriceBaseBuysFor = base_info.PriceBaseSellsFor
		}

		base_info.LevelRequired = market_good.LevelRequired.Get()
		base_info.RepRequired = market_good.RepRequired.Get()

		base_info.BaseInfo = e.GetBaseInfo(universe_mapped.BaseNickname(base_info.BaseNickname))

		if e.useful_bases_by_nick != nil {
			if _, ok := e.useful_bases_by_nick[base_info.BaseNickname]; !ok {
				base_info.NotBuyable = true
			}
		}

		goods_per_base[base_info.BaseNickname] = base_info
	}

	if e.configs.Discovery != nil {
		serverside_overrides := e.ServerSideMarketGoodsOverrides(commodity)
		for _, item := range serverside_overrides {
			goods_per_base[item.BaseNickname] = item
		}
	}

	return goods_per_base
}

type BaseInfo struct {
	BaseName    string
	SystemName  string
	Region      string
	FactionName string
	BasePos     cfgtype.Vector
	SectorCoord string
}

func (e *Exporter) GetRegionName(system *universe_mapped.System) string {
	var Region string
	system_infocard_Id := system.Ids_info.Get()
	if value, ok := e.configs.Infocards.Infocards[system_infocard_Id]; ok {
		if len(value.Lines) > 0 {
			Region = value.Lines[0]
		}
	}

	if strings.Contains(Region, "Sometimes limbo") && len(Region) > 11 {
		Region = Region[:20] + "..."
	}
	return Region
}

func (e *Exporter) GetBaseInfo(base_nickname universe_mapped.BaseNickname) BaseInfo {
	var result BaseInfo
	universe_base, found_universe_base := e.configs.Universe_config.BasesMap[universe_mapped.BaseNickname(base_nickname)]

	if !found_universe_base {
		return result
	}

	result.BaseName = e.GetInfocardName(universe_base.StridName.Get(), string(base_nickname))
	system_nickname := universe_base.System.Get()

	system, system_ok := e.configs.Universe_config.SystemMap[universe_mapped.SystemNickname(system_nickname)]
	if system_ok {
		result.SystemName = e.GetInfocardName(system.Strid_name.Get(), system_nickname)
		result.Region = e.GetRegionName(system)
	}

	var reputation_nickname string
	if system, ok := e.configs.Systems.SystemsMap[universe_base.System.Get()]; ok {
		for _, system_base := range system.Bases {
			if system_base.IdsName.Get() != universe_base.StridName.Get() {
				continue
			}

			reputation_nickname = system_base.RepNickname.Get()
			result.BasePos = system_base.Pos.Get()
		}

	}

	result.SectorCoord = VectorToSectorCoord(system, result.BasePos)

	var factionName string
	if group, exists := e.configs.InitialWorld.GroupsMap[reputation_nickname]; exists {
		factionName = e.GetInfocardName(group.IdsName.Get(), reputation_nickname)
	}

	result.FactionName = factionName

	return result
}

func (e *Exporter) FilterToUsefulCommodities(commodities []*Commodity) []*Commodity {
	var items []*Commodity = make([]*Commodity, 0, len(commodities))
	for _, item := range commodities {
		if !e.Buyable(item.Bases) {
			continue
		}
		items = append(items, item)
	}
	return items
}
