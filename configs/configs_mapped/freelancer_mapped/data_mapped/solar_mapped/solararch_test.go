package solar_mapped

import (
	"testing"

	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/iniload"
	"github.com/darklab8/fl-configs/configs/tests"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	fileref := tests.FixtureFileFind().GetFile(FILENAME)

	loaded_market_ships := Read(iniload.NewLoader(fileref).Scan())

	assert.Greater(t, len(loaded_market_ships.Solars), 0, "expected finding some elements")
}
