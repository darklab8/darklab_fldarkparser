/*
Tool to parse freelancer configs
*/
package configs_mapped

import (
	"sync"

	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/equipment_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/initialworld"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/interface_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/missions_mapped/empathy_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/ship_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/universe_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/universe_mapped/systems_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/exe_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/infocard_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/infocard_mapped/infocard"
	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/filefind"
	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/filefind/file"
	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/iniload"
	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/semantic"
	"github.com/darklab8/fl-configs/configs/settings/logus"

	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/equipment_mapped/equip_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/equipment_mapped/market_mapped"

	"github.com/darklab8/go-utils/goutils/utils"
	"github.com/darklab8/go-utils/goutils/utils/time_measure"
	"github.com/darklab8/go-utils/goutils/utils/utils_logus"
	"github.com/darklab8/go-utils/goutils/utils/utils_types"
)

type MappedConfigs struct {
	FreelancerINI *exe_mapped.Config

	Universe_config *universe_mapped.Config
	Systems         *systems_mapped.Config

	Market   *market_mapped.Config
	Equip    *equip_mapped.Config
	Goods    *equipment_mapped.Config
	Shiparch *ship_mapped.Config

	InfocardmapINI *interface_mapped.Config
	Infocards      *infocard.Config
	InitialWorld   *initialworld.Config
	Empathy        *empathy_mapped.Config
}

func NewMappedConfigs() *MappedConfigs {
	return &MappedConfigs{}
}

func getConfigs(filesystem *filefind.Filesystem, paths []*semantic.Path) []*iniload.IniLoader {
	return utils.CompL(paths, func(x *semantic.Path) *iniload.IniLoader {
		return iniload.NewLoader(filesystem.GetFile(utils_types.FilePath(x.FileName())))
	})
}

func (p *MappedConfigs) Read(file1path utils_types.FilePath) *MappedConfigs {
	logus.Log.Info("Parse START for FreelancerFolderLocation=", utils_logus.FilePath(file1path))
	filesystem := filefind.FindConfigs(file1path)
	p.FreelancerINI = exe_mapped.Read(iniload.NewLoader(filesystem.GetFile(exe_mapped.FILENAME_FL_INI)).Scan())

	files_goods := getConfigs(filesystem, p.FreelancerINI.Goods)
	files_market := getConfigs(filesystem, p.FreelancerINI.Markets)
	files_equip := getConfigs(filesystem, p.FreelancerINI.Equips)
	files_shiparch := getConfigs(filesystem, p.FreelancerINI.Ships)
	file_universe := iniload.NewLoader(filesystem.GetFile(universe_mapped.FILENAME))
	file_interface := iniload.NewLoader(filesystem.GetFile(interface_mapped.FILENAME_FL_INI))
	file_initialworld := iniload.NewLoader(filesystem.GetFile(initialworld.FILENAME))
	file_empathy := iniload.NewLoader(filesystem.GetFile(empathy_mapped.FILENAME))

	all_files := append(files_goods, files_market...)
	all_files = append(all_files, files_equip...)
	all_files = append(all_files, files_shiparch...)
	all_files = append(all_files,
		file_universe,
		file_interface,
		file_initialworld,
		file_empathy,
	)
	time_measure.TimeMeasure(func(m *time_measure.TimeMeasurer) {
		var wg sync.WaitGroup
		for _, file := range all_files {
			wg.Add(1)
			go func(file *iniload.IniLoader) {
				file.Scan()
				wg.Done()
			}(file)
		}
		wg.Wait()
	}, time_measure.WithMsg("Scanned ini loaders"))

	time_measure.TimeMeasure(func(m *time_measure.TimeMeasurer) {
		p.Universe_config = universe_mapped.Read(file_universe)

		p.Systems = systems_mapped.Read(p.Universe_config, filesystem)

		p.Market = market_mapped.Read(files_market)
		p.Equip = equip_mapped.Read(files_equip)
		p.Goods = equipment_mapped.Read(files_goods)
		p.Shiparch = ship_mapped.Read(files_shiparch)

		p.InfocardmapINI = interface_mapped.Read(file_interface)
		p.Infocards = infocard_mapped.Read(filesystem, p.FreelancerINI, filesystem.GetFile(infocard_mapped.FILENAME, infocard_mapped.FILENAME_FALLBACK))

		p.InitialWorld = initialworld.Read(file_initialworld)
		p.Empathy = empathy_mapped.Read(file_empathy)
	}, time_measure.WithMsg("Mapped stuff"))

	logus.Log.Info("Parse OK for FreelancerFolderLocation=", utils_logus.FilePath(file1path))

	return p
}

type IsDruRun bool

func (p *MappedConfigs) Write(is_dry_run IsDruRun) {
	// write
	files := []*file.File{}

	files = append(files,
		p.Universe_config.Write(),
	)
	files = append(files, p.Systems.Write()...)
	files = append(files, p.Market.Write()...)

	if is_dry_run {
		return
	}

	for _, file := range files {
		file.WriteLines()
	}
}
