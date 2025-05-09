package main

import (
	"net/url"
	"strings"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const TOKEN_PREFIX string = "edge-cache-token"
const TOKEN_PARAM string = "hdnts"

func main() {}
func init() {
	proxywasm.SetVMContext(&vmContext{})
}

// vmContext implements types.VMContext.
type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	tokenName   string
	tokenPrefix string
}

// httpContext implements types.HttpContext.
type httpContext struct {
	// Embed the default root http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	tokenName   string
	tokenPrefix string
}

// NewPluginContext implements types.VMContext.
func (vc *vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

// NewHttpContext implements types.PluginContext.
func (pc *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{tokenName: pc.tokenName, tokenPrefix: pc.tokenPrefix}
}

// OnPluginStart implements types.PluginContext.
func (pc *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	pc.tokenName = TOKEN_PARAM
	pc.tokenPrefix = TOKEN_PREFIX

	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogWarnf("no plugin configuration provided, using defaults: %v", err)
	} else {
		pc.tokenName = string(data)
	}
	return types.OnPluginStartStatusOK
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	headers, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
	}
	proxywasm.LogInfof("request headers: %v", headers)

	// Check if the request URL ends with a mpd
	path, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		proxywasm.LogCriticalf("failed to get ':path' header: %v", err)
		return types.ActionContinue
	}

	// Try getting full request url from header
	requestUrl, err := proxywasm.GetHttpRequestHeader("x-client-request-url")
	if err != nil {
		proxywasm.LogWarnf("no 'x-client-request-url' header found, use path instead: %v", err)
		requestUrl = path
	}

	isRedirect := false
	newPath := ""

	req := strings.Split(requestUrl, "?")
	proxywasm.LogInfof("req: %v", req)
	if len(req) == 2 {
		// query params present
		params, err := url.ParseQuery(req[1])
		if err != nil {
			proxywasm.LogCriticalf("failed to parse query params: %v", err)
			return types.ActionContinue
		}

		tokenValue := params.Get(ctx.tokenName)
		proxywasm.LogInfof("tokenValue: %s", tokenValue)
		if tokenValue != "" {
			// Token is passed as query parameter here, we need to redirect to path
			newPath = "/" + TOKEN_PREFIX + "/" + tokenValue + req[0]
			isRedirect = true
		}
	} else {
		// query params not present, strip token from path
		segments := strings.Split(path, "/")
		proxywasm.LogInfof("segments: %v", segments)

		// Path starts with "/" the first element is an empty string
		if segments[1] == TOKEN_PREFIX {
			token := segments[2]

			// if token is in path and ends with .mpd, skip
			if strings.HasSuffix(path, ".mpd") {
				return types.ActionContinue
			}

			proxywasm.LogInfof("token: %s", token)
			newPath = "/" + strings.Join(segments[3:], "/") + "?" + ctx.tokenName + "=" + token
			isRedirect = true
		} else {
			err := proxywasm.SendHttpResponse(403, nil, []byte("Forbidden"), -1)
			if err != nil {
				proxywasm.LogCriticalf("failed to send response: %v", err)
				return types.ActionContinue
			}
		}
	}

	if isRedirect {
		// Send redirect response
		headers := [][2]string{{"Location", newPath}}
		body := []byte("Redirecting to " + newPath)
		proxywasm.LogInfof("redirecting to %s", newPath)
		err := proxywasm.SendHttpResponse(302, headers, body, -1)
		if err != nil {
			proxywasm.LogCriticalf("failed to send response: %v", err)
			return types.ActionContinue
		}
		return types.ActionPause
	}
	return types.ActionContinue
}
