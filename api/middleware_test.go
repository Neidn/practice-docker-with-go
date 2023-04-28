package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"practice-docker/token"
	"testing"
	"time"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	createToken, err := tokenMaker.CreateToken(username, duration)
	require.NoErrorf(t, err, "failed to create access token: %v", err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, createToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestServer_authMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "username", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorizationHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "Invalid", "username", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationType",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "username", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "username", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// loop through test cases
	for i := range testCases {
		tc := testCases[i]

		// start test server and send request
		server := newTestServer(t, nil)

		// setup temporary endpoint for testing
		url := fmt.Sprintf("/auth")
		server.router.GET(
			url,
			authMiddleware(server.tokenMaker),
			func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			},
		)

		// run subtest
		t.Run(tc.name, func(t *testing.T) {
			//server = newTestServer(t, nil)
			recorder := httptest.NewRecorder()

			// create a new http request
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoErrorf(t, err, "failed to create request: %v", err)

			tc.setupAuth(t, request, server.tokenMaker)

			// call the endpoint
			server.router.ServeHTTP(recorder, request)

			// check the response
			tc.checkResponse(t, recorder)
		})
	}
}
