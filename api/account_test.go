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
	mockdb "github.com/homocode/bank_demo/api/mock"
	db "github.com/homocode/bank_demo/db/sqlc"
	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

func TestCreateAccountApi(t *testing.T) {
	reqBody := createAccountRequest{
		Owner:    util.RandomOwner(),
		Currency: util.RandomCurrency(),
	}

	// this is to pass it to Return in the store.EXPECT() to match the type
	account := randomAccount("")

	testCases := []struct {
		name          string
		reqBody       createAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "Ok",
			reqBody: reqBody,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    reqBody.Owner,
					Currency: reqBody.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, account)
			},
		},
		{
			name: "Invalid Request",
			reqBody: createAccountRequest{
				Owner:    "",
				Currency: "something",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "Internal Server Error",
			reqBody: reqBody,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    reqBody.Owner,
					Currency: reqBody.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			// marshal body data to JSON
			jsonBody, _ := json.Marshal(tc.reqBody)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
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

	reqParams := getAccountRequest{
		Id: account.ID,
	}

	testCases := []struct {
		name          string
		reqParams     getAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Ok",
			reqParams: reqParams,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(reqParams.Id)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "Invalid Request",
			reqParams: getAccountRequest{Id: 0},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "Not Found",
			reqParams: getAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(reqParams.Id)).
					Times(1).
					Return(db.Accounts{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				requireBodyToMatachAccount(t, recorder.Body, db.Accounts{})
			},
		},
		{
			name:      "Internal Server Error",
			reqParams: reqParams,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			url := fmt.Sprintf("/accounts/%d", tc.reqParams.Id)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// check response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})

	}

}

func TestListAccountsAPI(t *testing.T) {
	owner := util.RandomOwner()
	pageSize := 10

	accounts := make([]db.Accounts, pageSize)
	for i := 0; i < pageSize; i++ {
		accounts[i] = randomAccount(owner)
	}

	fmt.Println(accounts)

	reqQuery := listAccountsRequest{
		Owner:    owner,
		PageId:   1,
		PageSize: int32(pageSize),
	}

	testCases := []struct {
		name          string
		reqQuery      listAccountsRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "Ok",
			reqQuery: reqQuery,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  reqQuery.Owner,
					Limit:  reqQuery.PageSize,
					Offset: (reqQuery.PageSize * reqQuery.PageId) - reqQuery.PageSize,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyToMatachAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "Invalid Request",
			reqQuery: listAccountsRequest{
				Owner:    "",
				PageId:   0,
				PageSize: 2,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "Not Found",
			reqQuery: reqQuery,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  reqQuery.Owner,
					Limit:  reqQuery.PageSize,
					Offset: (reqQuery.PageSize * reqQuery.PageId) - reqQuery.PageSize,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Accounts{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "Internal Server Error",
			reqQuery: reqQuery,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  reqQuery.Owner,
					Limit:  reqQuery.PageSize,
					Offset: (reqQuery.PageSize * reqQuery.PageId) - reqQuery.PageSize,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Accounts{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/accounts"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// add query parameters to request URL
			q := request.URL.Query()
			q.Add("owner", fmt.Sprintf("%v", tc.reqQuery.Owner))
			q.Add("page_id", fmt.Sprintf("%d", tc.reqQuery.PageId))
			q.Add("page_size", fmt.Sprintf("%d", tc.reqQuery.PageSize))
			request.URL.RawQuery = q.Encode()

			// check response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
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

func requireBodyToMatachAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Accounts) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Accounts

	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
