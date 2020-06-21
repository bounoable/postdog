package storetypes

import (
	types "github.com/bounoable/postdog/api/ptypes"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/api/storeproto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
)

// LetterProto encodes let.
func LetterProto(let store.Letter) (*storeproto.Letter, error) {
	sentAt, err := ptypes.TimestampProto(let.SentAt)
	if err != nil {
		return nil, err
	}

	return &storeproto.Letter{
		Id:        let.ID.String(),
		Letter:    types.LetterProto(let.Letter),
		SentAt:    sentAt,
		SendError: let.SendError,
	}, nil
}

// Letter decodes let.
func Letter(let *storeproto.Letter) (store.Letter, error) {
	id, err := uuid.Parse(let.GetId())
	if err != nil {
		return store.Letter{}, err
	}

	sentAt, err := ptypes.Timestamp(let.GetSentAt())
	if err != nil {
		return store.Letter{}, err
	}

	return store.Letter{
		ID:        id,
		Letter:    types.Letter(let.GetLetter()),
		SentAt:    sentAt,
		SendError: let.GetSendError(),
	}, nil
}

// UUIDProto encodes id.
func UUIDProto(id uuid.UUID) *storeproto.UUID {
	return &storeproto.UUID{
		Id: id.String(),
	}
}

// UUID decodes id.
func UUID(id *storeproto.UUID) (uuid.UUID, error) {
	return uuid.Parse(id.GetId())
}
