package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	mockDB "practice-docker/db/mock"
	db "practice-docker/db/sqlc"
	"practice-docker/token"
	"practice-docker/util"
	"testing"
	"time"
)

func randomAccount(owner string) db.Accounts {
	return db.Accounts{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Accounts) {
	data, err := io.ReadAll(body)
	require.NoErrorf(t, err, "cannot read response body: %v", err)

	var gotAccount db.Accounts
	err = json.Unmarshal(data, &gotAccount)
	require.NoErrorf(t, err, "cannot unmarshal response body: %v", err)

	require.Equalf(t, account, gotAccount, "account mismatch")
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Accounts) {
	data, err := io.ReadAll(body)
	require.NoErrorf(t, err, "cannot read response body: %v", err)

	var gotAccounts []db.Accounts
	err = json.Unmarshal(data, &gotAccounts)
	require.NoErrorf(t, err, "cannot unmarshal response body: %v", err)

	require.Equalf(t, accounts, gotAccounts, "account mismatch")
}

// TestServer_GetAccountAPI is a unit test for the GetAccountAPI handler.
// GET /accounts/:id
func TestServer_GetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// testCases is a slice of test cases for the unit tests.
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized", time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusUnauthorized, recorder.Code, "response code should be %d", http.StatusUnauthorized)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusUnauthorized, recorder.Code, "response code should be %d", http.StatusUnauthorized)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Accounts{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusNotFound, recorder.Code, "response code should be %d", http.StatusNotFound)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusInternalServerError, recorder.Code, "response code should be %d", http.StatusInternalServerError)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusBadRequest, recorder.Code, "response code should be %d", http.StatusBadRequest)
			},
		},
	}

	// loop each test case.
	for i := range testCases {
		tc := testCases[i]

		// run each test case.
		t.Run(tc.name, func(t *testing.T) {
			// create a mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create mock store
			store := mockDB.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoErrorf(t, err, "cannot create request: %v", err)

			// setup authorization
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}

// TestServer_CreateAccountAPI tests the API endpoint for creating a new account.
// POST /accounts
func TestServer_CreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				arg := db.CreateAccountsParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}

				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusUnauthorized, recorder.Code, "response code should be %d", http.StatusUnauthorized)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				arg := db.CreateAccountsParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}

				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusInternalServerError, recorder.Code, "response code should be %d", http.StatusInternalServerError)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"owner":    user.Username,
				"currency": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusBadRequest, recorder.Code, "response code should be %d", http.StatusBadRequest)
			},
		},
		{
			name: "ViolateUniqueConstraint",
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Accounts{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusForbidden, recorder.Code, "response code should be %d", http.StatusForbidden)
			},
		},
		{
			name: "ViolationOfForeignKey",
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					CreateAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Accounts{}, &pq.Error{Code: "23503"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusForbidden, recorder.Code, "response code should be %d", http.StatusForbidden)
			},
		},
	}

	// loop each test case.
	for i := range testCases {
		tc := testCases[i]

		// run each test case.
		t.Run(tc.name, func(t *testing.T) {
			// create a mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create mock store
			store := mockDB.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts")
			body, err := json.Marshal(tc.body)
			require.NoErrorf(t, err, "cannot marshal account: %v", err)

			// create request
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoErrorf(t, err, "cannot create request: %v", err)

			// set request header
			request.Header.Set("Content-Type", "application/json")

			// setup authentication
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}

// TestServer_ListAccountsAPI test list accounts API.
// GET /accounts?page_id=%d&page_size=%d
func TestServer_ListAccountsAPI(t *testing.T) {
	user, _ := randomUser(t)
	// random accounts
	n := 10
	accounts := make([]db.Accounts, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	testCases := []struct {
		name          string
		pageID        int
		pageSize      int
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageID:   1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name:     "Unauthenticated",
			pageID:   1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusUnauthorized, recorder.Code, "response code should be %d", http.StatusUnauthorized)
			},
		},
		{
			name:     "InternalError",
			pageID:   1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusInternalServerError, recorder.Code, "response code should be %d", http.StatusInternalServerError)
			},
		},
		{
			name:     "InvalidPage",
			pageID:   -1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusBadRequest, recorder.Code, "response code should be %d", http.StatusBadRequest)
			},
		},
		{
			name:     "InvalidPageSize",
			pageID:   1,
			pageSize: -1,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusBadRequest, recorder.Code, "response code should be %d", http.StatusBadRequest)
			},
		},
		{
			name:     "PageSizeTooLarge",
			pageID:   1,
			pageSize: 1000,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusBadRequest, recorder.Code, "response code should be %d", http.StatusBadRequest)
			},
		},
		{
			name:     "NoData",
			pageID:   1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Accounts{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchAccounts(t, recorder.Body, []db.Accounts{})
			},
		},
		{
			name:     "PartialData",
			pageID:   1,
			pageSize: n,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts[:3], nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchAccounts(t, recorder.Body, accounts[:3])
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// run each test case.
		t.Run(tc.name, func(t *testing.T) {
			// create a mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create mock store
			store := mockDB.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.pageID, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoErrorf(t, err, "cannot create request: %v", err)

			// setup authorization
			tc.setupAuth(t, request, server.tokenMaker)

			// setup request
			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}
