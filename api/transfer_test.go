package api

import (
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
	"time"
)

func randomTransfer() db.Transfers {
	return db.Transfers{
		ID:            util.RandomInt(1, 10),
		FromAccountID: util.RandomSQLint64(),
		ToAccountID:   util.RandomSQLint64(),
		Amount:        util.RandomMoney(),
		CreatedAt:     util.RandomTime(),
	}
}

func requireBodyMatchTransfer(t *testing.T, recorder *httptest.ResponseRecorder, transfer db.Transfers) {
	data, err := io.ReadAll(recorder.Body)

	require.NoErrorf(t, err, "cannot read response body: %v", err)

	var gotTransfer db.Transfers
	err = json.Unmarshal(data, &gotTransfer)
	require.NoErrorf(t, err, "cannot unmarshal response body: %v", err)

	require.Equalf(t, transfer.ID, gotTransfer.ID, "transfer id mismatch")
	require.Equalf(t, transfer.FromAccountID, gotTransfer.FromAccountID, "transfer from account id mismatch")
	require.Equalf(t, transfer.ToAccountID, gotTransfer.ToAccountID, "transfer to account id mismatch")
	require.Equalf(t, transfer.Amount, gotTransfer.Amount, "transfer amount mismatch")
	require.WithinDurationf(t, transfer.CreatedAt, gotTransfer.CreatedAt, time.Second, "transfer created at mismatch")
}

func TestServer_GetTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		transferID    int64
		fromAccountID sql.NullInt64
		toAccountID   sql.NullInt64
		buildStubs    func(store *mockDB.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			transferID:    transfer.ID,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(transfer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusOK, recorder.Code, "response code should be %d", http.StatusOK)
				requireBodyMatchTransfer(t, recorder, transfer)
			},
		},
		{
			name:          "NotFound",
			transferID:    transfer.ID,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			buildStubs: func(store *mockDB.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfers{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equalf(t, http.StatusNotFound, recorder.Code, "response code should be %d", http.StatusNotFound)
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

			// create a mock store
			store := mockDB.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/transfers/%d", transfer.ID)
			request := httptest.NewRequest("GET", url, nil)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(recorder)
		})
	}
}
