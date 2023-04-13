package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/homocode/bank_demo/db/mock"
	db "github.com/homocode/bank_demo/db/sqlc"
	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name          string
	reqInfo       interface{}
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(recorder *httptest.ResponseRecorder)
}

func TestCreateAccountApi(t *testing.T) {
	type reqBody struct {
		owner    string
		currency string
	}

	rB := reqBody{
		owner:    util.RandomOwner(),
		currency: util.RandomCurrency(),
	}

	// this is to pass it to Return in the store.EXPECT() to match the type
	account := randomAccount(rB.owner)

	testCases := []testCase{
		{
			name:    "Ok",
			reqInfo: rB,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    rB.owner,
					Balance:  0,
					Currency: rB.currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Println(">>", recorder)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, account)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build service stubs (simulations of the service)
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts/"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// check response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetAccountApi(t *testing.T) {
	owner := util.RandomOwner()
	account := randomAccount(owner)

	type reqParams struct{ accountId int64 }

	rP := reqParams{
		accountId: account.ID,
	}

	testCases := []struct {
		name          string
		reqParams     reqParams
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Ok",
			reqParams: rP,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(rP.accountId)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "InvalidId",
			reqParams: reqParams{accountId: 0},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			reqParams: reqParams{accountId: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(rP.accountId)).
					Times(1).
					Return(db.Accounts{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, db.Accounts{})
			},
		},
		{
			name:      "InternalServerError",
			reqParams: rP,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, db.Accounts{})
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build service stubs (simulations of the service)
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.reqParams.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// check response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}

}

func randomAccount(owner string) db.Accounts {
	return db.Accounts{
		ID:      util.RandomInt(1, 100),
		Owner:   owner,
		Balance: util.RandomMoney(),
	}
}

func requireBodyToMatachAccount(t *testing.T, body *bytes.Buffer, account db.Accounts) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Accounts

	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
