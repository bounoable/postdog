package storetypes

import (
	types "github.com/bounoable/postdog/api/ptypes"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/api/storeproto"
	"github.com/golang/protobuf/ptypes"
)

// LetterProto encodes let.
func LetterProto(let store.Letter) (*storeproto.Letter, error) {
	sentAt, err := ptypes.TimestampProto(let.SentAt)
	if err != nil {
		return nil, err
	}

	return &storeproto.Letter{
		Letter:    types.LetterProto(let.Letter),
		SentAt:    sentAt,
		SendError: let.SendError,
	}, nil
}

// Letter decodes let.
func Letter(let *storeproto.Letter) (store.Letter, error) {
	sentAt, err := ptypes.Timestamp(let.GetSentAt())
	if err != nil {
		return store.Letter{}, err
	}

	return store.Letter{
		Letter:    types.Letter(let.GetLetter()),
		SentAt:    sentAt,
		SendError: let.GetSendError(),
	}, nil
}
