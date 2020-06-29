package storetypes

import (
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/api/storeproto"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/golang/protobuf/ptypes"
)

// QueryProto encodes q.
func QueryProto(q query.Query) (*storeproto.Query, error) {
	before, err := ptypes.TimestampProto(q.SentAt.Before)
	if err != nil {
		return nil, err
	}

	after, err := ptypes.TimestampProto(q.SentAt.After)
	if err != nil {
		return nil, err
	}

	return &storeproto.Query{
		SentAt: &storeproto.SentAtFilter{
			Before: before,
			After:  after,
		},
		Subjects: q.Subjects,
		From:     q.From,
		To:       q.To,
		Cc:       q.CC,
		Bcc:      q.BCC,
		Attachment: &storeproto.AttachmentFilter{
			Names:        q.Attachment.Names,
			ContentTypes: q.Attachment.ContentTypes,
			Size: &storeproto.AttachmentSizeFilter{
				Exact:  toInt64Slice(q.Attachment.Size.Exact),
				Ranges: toProtoSizeRanges(q.Attachment.Size.Ranges),
			},
		},
		Sort: &storeproto.SortConfig{
			SortBy:    storeproto.Sorting(q.Sort.SortBy),
			Direction: storeproto.SortDirection(q.Sort.Dir),
		},
		Paginate: &storeproto.PaginateConfig{
			Page:    int64(q.Paginate.Page),
			PerPage: int64(q.Paginate.PerPage),
		},
	}, nil
}

func toInt64Slice(nums []int) []int64 {
	res := make([]int64, len(nums))
	for i, num := range nums {
		res[i] = int64(num)
	}
	return res
}

func toProtoSizeRanges(ranges [][2]int) []*storeproto.AttachmentSizeRange {
	res := make([]*storeproto.AttachmentSizeRange, len(ranges))
	for i, rang := range ranges {
		res[i] = &storeproto.AttachmentSizeRange{
			Min: int64(rang[0]),
			Max: int64(rang[1]),
		}
	}
	return res
}

// QueryResultProto encodes letters.
func QueryResultProto(letters []store.Letter) (*storeproto.QueryResult, error) {
	res := make([]*storeproto.Letter, len(letters))
	var err error
	for i, let := range letters {
		if res[i], err = LetterProto(let); err != nil {
			return nil, err
		}
	}
	return &storeproto.QueryResult{
		Letters: res,
	}, nil
}

// Query decodes q.
func Query(q *storeproto.Query) (query.Query, error) {
	before, err := ptypes.Timestamp(q.GetSentAt().GetBefore())
	if err != nil {
		return query.Query{}, err
	}

	after, err := ptypes.Timestamp(q.GetSentAt().GetAfter())
	if err != nil {
		return query.Query{}, err
	}

	return query.Query{
		SentAt: query.SentAtFilter{
			Before: before,
			After:  after,
		},
		Subjects: q.GetSubjects(),
		From:     q.GetFrom(),
		To:       q.GetTo(),
		CC:       q.GetCc(),
		BCC:      q.GetBcc(),
		Attachment: query.AttachmentFilter{
			Names:        q.GetAttachment().GetNames(),
			ContentTypes: q.GetAttachment().GetContentTypes(),
			Size: query.AttachmentSizeFilter{
				Exact:  toIntSlice(q.GetAttachment().GetSize().GetExact()),
				Ranges: toSizeRanges(q.GetAttachment().GetSize().GetRanges()),
			},
		},
		Sort: query.SortConfig{
			SortBy: query.Sorting(q.GetSort().GetSortBy()),
			Dir:    query.SortDirection(q.GetSort().GetDirection()),
		},
		Paginate: query.PaginateConfig{
			Page:    int(q.GetPaginate().GetPage()),
			PerPage: int(q.GetPaginate().GetPerPage()),
		},
	}, nil
}

func toIntSlice(nums []int64) []int {
	res := make([]int, len(nums))
	for i, num := range nums {
		res[i] = int(num)
	}
	return res
}

func toSizeRanges(ranges []*storeproto.AttachmentSizeRange) [][2]int {
	res := make([][2]int, len(ranges))
	for i, rang := range ranges {
		res[i] = [2]int{int(rang.GetMin()), int(rang.GetMax())}
	}
	return res
}

// QueryResult decodes letters.
func QueryResult(res *storeproto.QueryResult) ([]store.Letter, error) {
	letters := make([]store.Letter, len(res.GetLetters()))
	var err error
	for i, let := range res.GetLetters() {
		if letters[i], err = Letter(let); err != nil {
			return nil, err
		}
	}
	return letters, nil
}
