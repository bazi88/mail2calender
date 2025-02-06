package usecase

import (
	"context"
	"encoding/base64"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMIMEParser_Parse(t *testing.T) {
	tests := []struct {
		name          string
		emailContent  string
		expectedEmail *ParsedEmail
		expectedError bool
	}{
		{
			name: "email_with_attachment",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Email with Attachment
Content-Type: multipart/mixed; boundary=boundary123

--boundary123
Content-Type: text/plain
Content-Transfer-Encoding: 7bit

Email body text

--boundary123
Content-Type: application/octet-stream
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="test.txt"

SGVsbG8gV29ybGQ=

--boundary123--`,
			expectedEmail: &ParsedEmail{
				From:        &mail.Address{Address: "sender@example.com"},
				To:          []*mail.Address{{Address: "recipient@example.com"}},
				Subject:     "Test Email with Attachment",
				TextContent: strings.TrimSpace("Email body text\n"),
				Attachments: []Attachment{
					{
						Filename:    "test.txt",
						Data:        []byte("Hello World"),
						ContentType: "application/octet-stream",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "email_with_multiple_attachments",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Email with Multiple Attachments
Content-Type: multipart/mixed; boundary=boundary123

--boundary123
Content-Type: text/plain
Content-Transfer-Encoding: 7bit

Email body text

--boundary123
Content-Type: application/pdf
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="test.pdf"

UERGIGNvbnRlbnQ=

--boundary123
Content-Type: image/jpeg
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="test.jpg"

SlBFRyBjb250ZW50

--boundary123--`,
			expectedEmail: &ParsedEmail{
				From:        &mail.Address{Address: "sender@example.com"},
				To:          []*mail.Address{{Address: "recipient@example.com"}},
				Subject:     "Test Email with Multiple Attachments",
				TextContent: strings.TrimSpace("Email body text\n"),
				Attachments: []Attachment{
					{
						Filename:    "test.pdf",
						Data:        []byte("PDF content"),
						ContentType: "application/pdf",
					},
					{
						Filename:    "test.jpg",
						Data:        []byte("JPEG content"),
						ContentType: "image/jpeg",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMIMEParser()
			email, err := parser.Parse(context.Background(), tt.emailContent)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, email)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, email)
			assert.Equal(t, tt.expectedEmail.From.Address, email.From.Address)
			for i, expectedTo := range tt.expectedEmail.To {
				assert.Equal(t, expectedTo.Address, email.To[i].Address)
			}
			assert.Equal(t, tt.expectedEmail.Subject, email.Subject)
			assert.Equal(t, tt.expectedEmail.TextContent, strings.TrimSpace(email.TextContent))

			assert.Equal(t, len(tt.expectedEmail.Attachments), len(email.Attachments))
			for i, expectedAttachment := range tt.expectedEmail.Attachments {
				assert.Equal(t, expectedAttachment.Filename, email.Attachments[i].Filename)
				assert.Equal(t, expectedAttachment.ContentType, email.Attachments[i].ContentType)

				// Decode base64 content
				decodedData, err := base64.StdEncoding.DecodeString(string(email.Attachments[i].Data))
				assert.NoError(t, err)
				assert.Equal(t, expectedAttachment.Data, decodedData)
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
