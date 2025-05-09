package main

import (
	"testing"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/stretchr/testify/require"
)

func Test_httpContext_OnHttpRequestHeaders(t *testing.T) {
	type testCase struct {
		name            string
		path            string
		requestURL      string
		tokenName       string
		expectedAction  types.Action
		expectedStatus  int
		expectedHeaders [][2]string
		expectedBody    string
	}

	testCases := []testCase{
		{
			name:           "redirect_with_token_in_query_params",
			path:           "/test.mpd",
			requestURL:     "/test.mpd?hdnts=test_token",
			tokenName:      "hdnts",
			expectedAction: types.ActionPause,
			expectedStatus: 302,
			expectedHeaders: [][2]string{
				{"Location", "/edge-cache-token/test_token/test.mpd"},
			},
			expectedBody: "Redirecting to /edge-cache-token/test_token/test.mpd",
		},
		{
			name:           "no_redirect_no_token",
			path:           "/test.mpd",
			requestURL:     "/test.mpd",
			tokenName:      "hdnts",
			expectedAction: types.ActionContinue,
			expectedStatus: 0, // No response sent
		},
		{
			name:           "redirect_with_token_in_path",
			path:           "/edge-cache-token/test_token/test_file",
			requestURL:     "/edge-cache-token/test_token/test_file",
			tokenName:      "hdnts",
			expectedAction: types.ActionPause,
			expectedStatus: 302,
			expectedHeaders: [][2]string{
				{"Location", "/test_file?hdnts=test_token"},
			},
			expectedBody: "Redirecting to /test_file?hdnts=test_token",
		},
		{
			name:           "no_redirect_mpd_with_token_in_path",
			path:           "/edge-cache-token/test_token/test.mpd",
			requestURL:     "/edge-cache-token/test_token/test.mpd",
			tokenName:      "hdnts",
			expectedAction: types.ActionContinue,
			expectedStatus: 0,
		},
		{
			name:           "forbidden_no_token_in_path",
			path:           "/some_other_path",
			requestURL:     "/some_other_path",
			tokenName:      "hdnts",
			expectedAction: types.ActionContinue,
			expectedStatus: 403, expectedHeaders: [][2]string{},
			expectedBody: "Forbidden",
		},
		{
			name:           "custom_token_name",
			path:           "/test.mpd",
			requestURL:     "/test.mpd?custom_token=custom_value",
			tokenName:      "custom_token",
			expectedAction: types.ActionPause,
			expectedStatus: 302,
			expectedHeaders: [][2]string{
				{"Location", "/edge-cache-token/custom_value/test.mpd"},
			},
			expectedBody: "Redirecting to /edge-cache-token/custom_value/test.mpd",
		},
		{
			name:           "no_request_url_header",
			path:           "/test.mpd?hdnts=test_token",
			requestURL:     "", // Simulate missing header
			tokenName:      "hdnts",
			expectedAction: types.ActionPause,
			expectedStatus: 302,
			expectedHeaders: [][2]string{
				{"Location", "/edge-cache-token/test_token/test.mpd"},
			},
			expectedBody: "Redirecting to /edge-cache-token/test_token/test.mpd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			opt := proxytest.NewEmulatorOption().WithVMContext(&vmContext{}).WithPluginConfiguration([]byte(tc.tokenName))
			host, reset := proxytest.NewHostEmulator(opt)
			defer reset()

			require.Equal(t, types.OnPluginStartStatusOK, host.StartPlugin())

			// Create http context
			id := host.InitializeHttpContext()

			headers := [][2]string{
				{":method", "GET"},
				{":path", tc.path},
			}
			if tc.requestURL != "" {
				headers = append(headers, [2]string{"x-client-request-url", tc.requestURL})
			}

			action := host.CallOnRequestHeaders(id, headers, false)
			require.Equal(t, tc.expectedAction, action)

			if tc.expectedStatus != 0 {
				localResponse := host.GetSentLocalResponse(id)
				// Check response
				require.NotNil(t, localResponse)
				require.Equal(t, tc.expectedStatus, int(localResponse.StatusCode))
				respHeaders := localResponse.Headers
				require.Equal(t, tc.expectedHeaders, respHeaders)
				body := localResponse.Data
				require.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}
