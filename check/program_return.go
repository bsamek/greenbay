package check

import (
	"strings"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/registry"
	"github.com/pkg/errors"
)

func registerProgramReturnChecks() {
	programCheckFactoryFactory := func(name string, c compiler) func() amboy.Job {
		return func() amboy.Job {
			return &programReturnCheck{
				Base:     NewBase(name, 0),
				compiler: c,
			}
		}
	}

	registrar := func(table map[string]compilerFactory) {
		for name, factory := range table {
			if strings.Contains(name, "-script") {
				name = strings.Replace(name, "-script", "-script-no-output-check", 1)
				registry.AddJobType(name, programCheckFactoryFactory(name, factory()))
			}
		}
	}

	registrar(scriptCompilerInterfaceFactoryTable())
}

type programReturnCheck struct {
	Source   string `bson:"source" json:"source" yaml:"source"`
	*Base    `bson:"metadata" json:"metadata" yaml:"metadata"`
	compiler compiler
}

func (c *programReturnCheck) Run() {
	c.startTask()
	defer c.MarkComplete()

	c.setState(true)

	if err := c.compiler.Validate(); err != nil {
		c.setState(false)
		c.AddError(err)
		return
	}

	_, err := c.compiler.CompileAndRun(c.Source)
	if err != nil {
		c.setState(false)
		c.AddError(errors.New("program did not exit 0"))
		c.setMessage([]string{
			err.Error(),
		})
		return
	}
}
