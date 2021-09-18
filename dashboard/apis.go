package dashboard

import (
	"encoding/json"

	"github.com/blockpilabs/near-drpc/near"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Router struct {
	near *near.Near
	router *router.Router
}

func (r *Router)Nodes(ctx *fasthttp.RequestCtx) {
	result := make(map[string]interface{})
	result["code"] = "SUCCESS"
	result["data"] = r.near.Summary()
	data,_ := json.Marshal(result)

	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	ctx.Response.Header.SetContentType("application/json")
	ctx.Write(data)
}


func NewRouter(near *near.Near) *Router {
	r := &Router{
		near: near,
		router: router.New(),
	}
	r.router.GET("/api/status", r.Nodes)
	return r
}

