package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MIMEParser handles parsing of email content
type MIMEParser interface {
	Parse(ctx context.Context, emailContent string) (*ParsedEmail, error)
}

// ParsedEmail represents the parsed content of an email
type ParsedEmail struct {
	Subject     string
	From        *mail.Address
	To          []*mail.Address
	Cc          []*mail.Address
	TextContent string
	HTMLContent string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type mimeParserImpl struct {
	tracer trace.Tracer
}

// NewMIMEParser creates a new instance of MIMEParser
func NewMIMEParser() MIMEParser {
	return &mimeParserImpl{
		tracer: otel.Tracer("mime-parser"),
	}
}

func (p *mimeParserImpl) Parse(ctx context.Context, emailContent string) (*ParsedEmail, error) {
	_, span := p.tracer.Start(ctx, "ParseEmail")
	defer span.End()

	msg, err := mail.ReadMessage(strings.NewReader(emailContent))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read email: %v", err)
	}

	parsed := &ParsedEmail{}

	// Parse headers
	if err := p.parseHeaders(msg, parsed); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse headers: %v", err)
	}

	// Parse body
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		mediaType = "text/plain" // Default to text/plain if header is missing
	}

	span.SetAttributes(
		attribute.String("content.type", mediaType),
		attribute.String("subject", parsed.Subject),
	)

	if strings.HasPrefix(mediaType, "multipart/") {
		if err := p.parseMultipart(msg.Body, params["boundary"], parsed); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse multipart: %v", err)
		}
	} else {
		body, err := p.parseTextPart(msg.Body, msg.Header)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse body: %v", err)
		}
		if strings.HasPrefix(mediaType, "text/html") {
			parsed.HTMLContent = body
		} else {
			parsed.TextContent = body
		}
	}

	return parsed, nil
}

func (p *mimeParserImpl) parseHeaders(msg *mail.Message, parsed *ParsedEmail) error {
	// Parse Subject
	parsed.Subject = p.decodeHeader(msg.Header.Get("Subject"))

	// Parse From
	from, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		return fmt.Errorf("invalid From address: %v", err)
	}
	parsed.From = from

	// Parse To
	to, err := mail.ParseAddressList(msg.Header.Get("To"))
	if err == nil {
		parsed.To = to
	}

	// Parse Cc
	cc, err := mail.ParseAddressList(msg.Header.Get("Cc"))
	if err == nil {
		parsed.Cc = cc
	}

	return nil
}

func (p *mimeParserImpl) parseMultipart(r io.Reader, boundary string, parsed *ParsedEmail) error {
	multipartReader := multipart.NewReader(r, boundary)

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		contentType := part.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			continue
		}

		if strings.HasPrefix(mediaType, "text/") {
			content, err := p.parseTextPartFromHeader(part, part.Header)
			if err != nil {
				continue
			}

			if mediaType == "text/html" {
				parsed.HTMLContent = content
			} else {
				parsed.TextContent = content
			}
		} else {
			// Handle attachment
			if err := p.parseAttachment(part, parsed); err != nil {
				continue
			}
		}
	}

	return nil
}

func (p *mimeParserImpl) parseTextPart(r io.Reader, header mail.Header) (string, error) {
	contentType := header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		params = make(map[string]string)
	}

	return p.parseTextContent(r, header.Get("Content-Transfer-Encoding"), params["charset"])
}

func (p *mimeParserImpl) parseTextPartFromHeader(r io.Reader, header textproto.MIMEHeader) (string, error) {
	contentType := header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		params = make(map[string]string)
	}

	return p.parseTextContent(r, header.Get("Content-Transfer-Encoding"), params["charset"])
}

func (p *mimeParserImpl) parseTextContent(r io.Reader, transferEncoding string, charset string) (string, error) {
	if charset == "" {
		charset = "utf-8"
	}

	var contentReader io.Reader = r

	// Apply transfer encoding
	switch strings.ToLower(transferEncoding) {
	case "base64":
		contentReader = base64.NewDecoder(base64.StdEncoding, contentReader)
	case "quoted-printable":
		contentReader = quotedprintable.NewReader(contentReader)
	}

	// Apply character encoding
	if dec := p.getDecoder(charset); dec != nil {
		contentReader = transform.NewReader(contentReader, dec.NewDecoder())
	}

	content, err := io.ReadAll(contentReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (p *mimeParserImpl) parseAttachment(part *multipart.Part, parsed *ParsedEmail) error {
	filename := p.decodeHeader(part.FileName())
	if filename == "" {
		return nil
	}

	data, err := io.ReadAll(part)
	if err != nil {
		return err
	}

	parsed.Attachments = append(parsed.Attachments, Attachment{
		Filename:    filename,
		ContentType: part.Header.Get("Content-Type"),
		Data:        data,
	})

	return nil
}

func (p *mimeParserImpl) decodeHeader(header string) string {
	decoded, err := (&mime.WordDecoder{}).DecodeHeader(header)
	if err != nil {
		return header
	}
	return decoded
}

func (p *mimeParserImpl) getDecoder(charset string) encoding.Encoding {
	switch strings.ToLower(charset) {
	case "windows-1252":
		return charmap.Windows1252
	case "iso-8859-1":
		return charmap.ISO8859_1
	case "shift-jis":
		return japanese.ShiftJIS
	case "euc-jp":
		return japanese.EUCJP
	case "euc-kr":
		return korean.EUCKR
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	default:
		return nil // UTF-8 or unknown
	}
}
