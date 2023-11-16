package option

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/pflag"
)

type Common struct {
	Debug    bool
	Verbose  bool
	NoColors bool
	Colors
}

type Colors struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
}

func (opts *Common) ApplyFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&opts.Debug, "debug", false, "debug mode")
	fs.BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	fs.BoolVar(&opts.NoColors, "no-color", false, "If specified, output won't contain any color")
}

func (opts *Common) Parse() error {
	if opts.NoColors {
		lipgloss.SetColorProfile(termenv.Ascii)
		return nil
	}

	opts.Colors = Colors{
		Primary:   lipgloss.Color("#EF4136"),
		Secondary: lipgloss.Color("#26A568"),
	}

	return nil
}
