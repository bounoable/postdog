package ptypes

import (
	"net/mail"
	"net/textproto"

	"github.com/bounoable/postdog/api/proto"
	"github.com/bounoable/postdog/letter"
)

// LetterProto encodes let.
func LetterProto(let letter.Letter) *proto.Letter {
	return &proto.Letter{
		Subject:     let.Subject,
		From:        AddressProto(let.From),
		To:          AddressProtos(let.To...),
		Cc:          AddressProtos(let.CC...),
		Bcc:         AddressProtos(let.BCC...),
		ReplyTo:     AddressProtos(let.ReplyTo...),
		Text:        let.Text,
		Html:        let.HTML,
		Attachments: AttachmentProtos(let.Attachments...),
	}
}

// AddressProto encodes addr.
func AddressProto(addr mail.Address) *proto.Address {
	return &proto.Address{
		Name:    addr.Name,
		Address: addr.Address,
	}
}

// AddressProtos encodes addrs.
func AddressProtos(addrs ...mail.Address) []*proto.Address {
	res := make([]*proto.Address, len(addrs))
	for i, addr := range addrs {
		res[i] = AddressProto(addr)
	}
	return res
}

// AttachmentProto encodes attach.
func AttachmentProto(attach letter.Attachment) *proto.Attachment {
	return &proto.Attachment{
		Filename: attach.Filename,
		Header:   AttachmentHeaderProto(attach.Header),
		Content:  attach.Content,
	}
}

// AttachmentHeaderProto encodes header.
func AttachmentHeaderProto(header textproto.MIMEHeader) map[string]*proto.HeaderValues {
	pbheader := make(map[string]*proto.HeaderValues)
	for key, vals := range header {
		pbheader[key] = &proto.HeaderValues{
			Values: vals,
		}
	}
	return pbheader
}

// AttachmentProtos encodes attachs.
func AttachmentProtos(attachs ...letter.Attachment) []*proto.Attachment {
	res := make([]*proto.Attachment, len(attachs))
	for i, attach := range attachs {
		res[i] = AttachmentProto(attach)
	}
	return res
}

// Letter decodes let.
func Letter(let *proto.Letter) letter.Letter {
	return letter.Letter{
		Subject:     let.GetSubject(),
		From:        Address(let.GetFrom()),
		To:          Addresses(let.GetTo()...),
		CC:          Addresses(let.GetCc()...),
		BCC:         Addresses(let.GetBcc()...),
		ReplyTo:     Addresses(let.GetReplyTo()...),
		Text:        let.GetText(),
		HTML:        let.GetHtml(),
		Attachments: Attachments(let.GetAttachments()...),
	}
}

// Address decodes addr.
func Address(addr *proto.Address) mail.Address {
	return mail.Address{
		Name:    addr.GetName(),
		Address: addr.GetAddress(),
	}
}

// Addresses decodes addrs.
func Addresses(addrs ...*proto.Address) []mail.Address {
	res := make([]mail.Address, len(addrs))
	for i, addr := range addrs {
		res[i] = Address(addr)
	}
	return res
}

// Attachment decodes attach.
func Attachment(attach *proto.Attachment) letter.Attachment {
	return letter.Attachment{
		Filename: attach.GetFilename(),
		Header:   AttachmentHeader(attach.GetHeader()),
		Content:  attach.GetContent(),
	}
}

// AttachmentHeader decodes header.
func AttachmentHeader(header map[string]*proto.HeaderValues) textproto.MIMEHeader {
	res := make(textproto.MIMEHeader)
	for key, vals := range header {
		for _, val := range vals.GetValues() {
			res.Add(key, val)
		}
	}
	return res
}

// Attachments decodes attachs.
func Attachments(attachs ...*proto.Attachment) []letter.Attachment {
	res := make([]letter.Attachment, len(attachs))
	for i, attach := range attachs {
		res[i] = Attachment(attach)
	}
	return res
}
