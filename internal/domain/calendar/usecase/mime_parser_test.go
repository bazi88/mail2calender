package usecase

import (
	"context"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMIMEParser_Parse(t *testing.T) {
	tests := []struct {
		name         string
		emailContent string
		want         *ParsedEmail
		wantErr      bool
	}{
		{
			name: "simple plain text email",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Subject
Content-Type: text/plain

This is a test email body.`,
			want: &ParsedEmail{
				Subject:     "Test Subject",
				From:        &mail.Address{Address: "sender@example.com"},
				To:          []*mail.Address{{Address: "recipient@example.com"}},
				TextContent: "This is a test email body.",
			},
			wantErr: false,
		},
		{
			name: "multipart email with HTML and text",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Multipart
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain

Plain text content
--boundary123
Content-Type: text/html

<html><body>HTML content</body></html>
--boundary123--`,
			want: &ParsedEmail{
				Subject:     "Test Multipart",
				From:        &mail.Address{Address: "sender@example.com"},
				To:          []*mail.Address{{Address: "recipient@example.com"}},
				TextContent: "Plain text content",
				HTMLContent: "<html><body>HTML content</body></html>",
			},
			wantErr: false,
		},
		{
			name: "email with attachment",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Attachment
Content-Type: multipart/mixed; boundary="boundary123"

--boundary123
Content-Type: text/plain

Email with attachment
--boundary123
Content-Type: application/pdf
Content-Disposition: attachment; filename="test.pdf"
Content-Transfer-Encoding: base64

SGVsbG8gV29ybGQ=
--boundary123--`,
			want: &ParsedEmail{
				Subject:     "Test Attachment",
				From:        &mail.Address{Address: "sender@example.com"},
				To:          []*mail.Address{{Address: "recipient@example.com"}},
				TextContent: "Email with attachment",
				Attachments: []Attachment{
					{
						Filename:    "test.pdf",
						ContentType: "application/pdf",
						Data:        []byte("Hello World"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "invalid email format",
			emailContent: "invalid email content",
			want:         nil,
			wantErr:      true,
		},
	}

	parser := NewMIMEParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(context.Background(), tt.emailContent)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Subject, got.Subject)
			assert.Equal(t, tt.want.From.Address, got.From.Address)
			if len(tt.want.To) > 0 {
				assert.Equal(t, tt.want.To[0].Address, got.To[0].Address)
			}
			assert.Equal(t, tt.want.TextContent, got.TextContent)
			assert.Equal(t, tt.want.HTMLContent, got.HTMLContent)

			if len(tt.want.Attachments) > 0 {
				assert.Equal(t, len(tt.want.Attachments), len(got.Attachments))
				assert.Equal(t, tt.want.Attachments[0].Filename, got.Attachments[0].Filename)
				assert.Equal(t, tt.want.Attachments[0].ContentType, got.Attachments[0].ContentType)
				assert.Equal(t, tt.want.Attachments[0].Data, got.Attachments[0].Data)
			}
		})
	}
}

func TestMIMEParser_ParseHeaders(t *testing.T) {
	tests := []struct {
		name         string
		emailContent string
		wantSubject  string
		wantFrom     string
		wantTo       []string
		wantCc       []string
		wantErr      bool
	}{
		{
			name: "valid headers",
			emailContent: `From: sender@example.com
To: recipient1@example.com, recipient2@example.com
Cc: cc1@example.com, cc2@example.com
Subject: Test Headers
Content-Type: text/plain

Body content`,
			wantSubject: "Test Headers",
			wantFrom:    "sender@example.com",
			wantTo:      []string{"recipient1@example.com", "recipient2@example.com"},
			wantCc:      []string{"cc1@example.com", "cc2@example.com"},
			wantErr:     false,
		},
		{
			name: "missing from header",
			emailContent: `To: recipient@example.com
Subject: Test

Body`,
			wantErr: true,
		},
	}

	parser := NewMIMEParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(context.Background(), tt.emailContent)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSubject, got.Subject)
			assert.Equal(t, tt.wantFrom, got.From.Address)

			for i, addr := range tt.wantTo {
				assert.Equal(t, addr, got.To[i].Address)
			}

			for i, addr := range tt.wantCc {
				assert.Equal(t, addr, got.Cc[i].Address)
			}
		})
	}
}
