package demo

import "github.com/host-uk/core"

// this instance is the singleton instance of the demo module.
var instance *API

type API struct {
	name string
	core *core.Core
}

func main() {
	coreService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)

	rickService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)
	mortyService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)

	core.Mod[API](coreService, "demo").name = "demo"
	core.Mod[API](rickService).name = "demo2"
	core.Mod[API](mortyService).name = "demo2"

}

func RegisterDemo(c *core.Core) error {
	instance = &API{
		core: c,
	}
	if err := c.RegisterModule("demo", instance); err != nil {
		return err
	}
	c.RegisterAction(handleActionCall)
	return nil
}

func RegisterDemo2(c *core.Core) error {
	instance = &API{
		core: c,
	}
	if err := c.RegisterModule("demo", instance); err != nil {
		return err
	}
	c.RegisterAction(handleActionCall)
	return nil
}

func handleActionCall(c *core.Core, msg core.Message) error {
	return nil
}
