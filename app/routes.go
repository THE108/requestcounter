package app

import (
	"github.com/THE108/requestcounter/handlers/requestcount"
)

func (this *Application) getHandlers() []*HandlerInfo {
	return []*HandlerInfo{
		{
			Name:    "GetRequestCount",
			Method:  GET,
			Route:   "/requestcount",
			Handler: requestcount.NewGetRecipeHandler(this.models.requestCounter),
		},
	}
}
