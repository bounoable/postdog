package query_test

// func TestRun(t *testing.T) {
// 	letters := []letter.Letter{
// 		letter.Write(
// 			letter.Subject("Letter 1"),
// 			letter.MustAttach(bytes.NewReader([]byte{1, 2, 3}), "attachment.txt", letter.ContentType("text/plain")),
// 		),
// 		letter.Write(
// 			letter.Subject("Letter 2"),
// 			letter.MustAttach(bytes.NewReader([]byte{2, 3, 4}), "attachment.txt", letter.ContentType("text/plain")),
// 		),
// 	}

// 	cases := map[string]struct {
// 		configureStore  func(*mock_store.MockStore, *gomock.Controller)
// 		expectedLetters []letter.Letter
// 	}{
// 		"default query (query all)": {
// 			expectedLetters: letters,
// 		},
// 	}

// 	for name, tcase := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			store := mock_store.NewMockStore(ctrl)

// 			if tcase.configureStore != nil {
// 				tcase.configureStore(store, ctrl)
// 			}

// 			cur, err := query.Run(
// 				context.Background(),
// 				store,
// 			)
// 		})
// 	}
// }
