package sepa

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "sepa",
	Mode: core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsSepa{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsSepa); ok {
			return &scriptSepa{
				name:    name,
				baseURL: "https://apps.sepa.org.uk/database/riverlevels",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsSepa{})
	},
}
