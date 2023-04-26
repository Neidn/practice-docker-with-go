package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	mockDB "practice-docker/db/mock"
	db "practice-docker/db/sqlc"
	"practice-docker/util"
	"testing"
)

func randomAccount() db.Accounts {
	return db.Accounts{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
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

func TestServer_GetAccountAPI(t *testing.T) {
	account := randomAccount()

	// testCases is a slice of test cases for the unit tests.
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
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
			name:      "NotFound",
			accountID: account.ID,
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
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoErrorf(t, err, "cannot create request: %v", err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}

func TestServer_CreateAccountAPI(t *testing.T) {

}

func TestServer_ListAccountsAPI(t *testing.T) {
	// random accounts
	n := 10
	accounts := make([]db.Accounts, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount()
	}

	testCases := []struct {
		name          string
		pageID        int
		pageSize      int
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageID:   1,
			pageSize: n,
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
			name:     "InternalError",
			pageID:   1,
			pageSize: n,
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
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.pageID, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoErrorf(t, err, "cannot create request: %v", err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}
