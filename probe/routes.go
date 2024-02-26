package probe

import (
	"github.com/kataras/iris/v12/core/router"
	"github.com/zerok-ai/zk-operator/probe/handler"
)

func Initialize(app router.Party, ph handler.ProbeHandler) {

	probeAPI := app.Party("/p/probe")
	{
		probeAPI.Get("/", ph.GetAllProbes)
		probeAPI.Post("/", ph.CreateProbe)
		probeAPI.Delete("/{name}", ph.DeleteProbe)
		probeAPI.Put("/{name}", ph.UpdateProbe)
		probeAPI.Get("/service", ph.GetAllServices)
	}
}
