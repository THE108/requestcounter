package requestcount

import (
	"github.com/THE108/requestcounter/models/requestcount"
	"github.com/THE108/requestcounter/utils/params"

	"golang.org/x/net/context"
)

type IRequestCountGetter interface {
	Get(ctx context.Context) *requestcount.RequestCount
}

type GetRequestCountHandler struct {
	model IRequestCountGetter
}

func NewGetRecipeHandler(model IRequestCountGetter) *GetRequestCountHandler {
	return &GetRequestCountHandler{
		model: model,
	}
}

func (handler *GetRequestCountHandler) Process(ctx context.Context, _ params.Params) (interface{}, error) {
	return handler.model.Get(ctx), nil
}
