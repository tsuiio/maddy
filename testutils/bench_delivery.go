package testutils

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/emersion/go-message/textproto"
	"github.com/foxcpp/maddy/buffer"
	"github.com/foxcpp/maddy/module"
)

// Empirically observed "around average" values.
const (
	MessageBodySize             = 100 * 1024
	ExtraMessageHeaderFields    = 10
	ExtraMessageHeaderFieldSize = 50
)

const testHeaderString = "Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.org>\r\n" +
	"Reply-To: Mitsuha Miyamizu <mitsuha.miyamizu+replyto@example.org>\r\n" +
	"Message-Id: 42@example.org\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Transfer-Encoding: 8but\r\n" +
	"Subject: Your Name.\r\n" +
	"To: Taki Tachibana <taki.tachibana@example.org>\r\n" +
	"\r\n"

const testHeaderFromToString = "From: Mitsuha Miyamizu <mitsuha.miyamizu@example.org>\r\n" +
	"To: Taki Tachibana <taki.tachibana@example.org>\r\n" +
	"\r\n"

const testHeaderDateString = "Date: Sat, 18 Jun 2016 12:00:00 +0900\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"\r\n"

const testHeaderNoFromToString = "Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
	"Date: Sat, 18 Jun 2016 12:00:00 +0900\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"Reply-To: Mitsuha Miyamizu <mitsuha.miyamizu+replyto@example.org>\r\n" +
	"Message-Id: 42@example.org\r\n" +
	"Subject: Your Name.\r\n" +
	"\r\n"

const testAltHeaderString = "Content-Type: multipart/alternative; boundary=b2\r\n" +
	"\r\n"

const testTextHeaderString = "Content-Disposition: inline\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n"

const testTextContentTypeString = "Content-Type: text/plain\r\n" +
	"\r\n"

const testTextNoContentTypeString = "Content-Disposition: inline\r\n" +
	"\r\n"

const testTextBodyString = "What's your name?"

const testTextString = testTextHeaderString + testTextBodyString

const testHTMLHeaderString = "Content-Disposition: inline\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n"

const testHTMLBodyString = "<div>What's <i>your</i> name?</div>"

const testHTMLString = testHTMLHeaderString + testHTMLBodyString

const testAttachmentHeaderString = "Content-Disposition: attachment; filename=note.txt\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n"

const testAttachmentBodyString = "My name is Mitsuha."

const testAttachmentString = testAttachmentHeaderString + testAttachmentBodyString

const testBodyString = "--message-boundary\r\n" +
	testAltHeaderString +
	"\r\n--b2\r\n" +
	testTextString +
	"\r\n--b2\r\n" +
	testHTMLString +
	"\r\n--b2--\r\n" +
	"\r\n--message-boundary\r\n" +
	testAttachmentString +
	"\r\n--message-boundary--\r\n"

var testMailString = testHeaderString + testBodyString + strings.Repeat("A", MessageBodySize)

type multipleErrs map[string]error

func (m multipleErrs) SetStatus(rcptTo string, err error) {
	m[rcptTo] = err
}

func RandomMsg(b *testing.B) (module.MsgMetadata, textproto.Header, buffer.Buffer) {
	IDRaw := sha1.Sum([]byte(b.Name()))
	encodedID := hex.EncodeToString(IDRaw[:])

	body := bufio.NewReader(strings.NewReader(testMailString))
	hdr, _ := textproto.ReadHeader(body)
	for i := 0; i < ExtraMessageHeaderFields; i++ {
		hdr.Add("AAAAAAAAAAAA-"+strconv.Itoa(i), strings.Repeat("A", ExtraMessageHeaderFieldSize))
	}
	bodyBlob, _ := ioutil.ReadAll(body)

	return module.MsgMetadata{
		DontTraceSender: true,
		ID:              encodedID,
	}, hdr, buffer.MemoryBuffer{Slice: bodyBlob}
}

func BenchDelivery(b *testing.B, target module.DeliveryTarget, sender string, recipientTemplates []string) {
	meta, header, body := RandomMsg(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delivery, err := target.Start(&meta, sender)
		if err != nil {
			b.Fatal(err)
		}

		for i, rcptTemplate := range recipientTemplates {
			rcpt := strings.Replace(rcptTemplate, "X", strconv.Itoa(i), -1)

			if err := delivery.AddRcpt(rcpt); err != nil {
				b.Fatal(err)
			}
		}

		if err := delivery.Body(header, body); err != nil {
			b.Fatal(err)
		}

		if err := delivery.Commit(); err != nil {
			b.Fatal(err)
		}
	}
}
