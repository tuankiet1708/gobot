package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
)

const (
	ExchangeRateVCBURL = "http://www.vietcombank.com.vn/ExchangeRates/ExrateXML.aspx"
)

type (
	ExchangeRate struct {
		DateTime string   `xml:"DateTime"`
		Exrate   []Exrate `xml:"Exrate"`
		Source   string   `xml:"Source"`
	}

	Exrate struct {
		CurrencyCode string `xml:"CurrencyCode,attr"`
		CurrencyName string `xml:"CurrencyName,attr"`
		Buy          string `xml:"Buy,attr"`
		Transfer     string `xml:"Transfer,attr"`
		Sell         string `xml:"Sell,attr"`
	}
)

var (
	// Lưu thông tin ngoại tệ lấy được
	exRateList *ExchangeRate
	// Lưu nhóm ngoại tệ đang hiển thị của từng người dùng
	exRateGroupMap = make(map[string]int)
)

// Xử lý dữ liệu với tỉ giá
/**
Hàm processMessage cung cấp chức năng về thông tin tỉ giá.

Đầu tiên chúng ta cho người dùng chọn ngoại tệ họ muốn xem tỉ giá. Có nhiều cách làm việc này:

1. Hiển thị danh sách các ngoại tệ, người dùng gõ viết tắt ngoại tệ nào thì bot trả lời thông tin tỉ giá ngoại tệ đó. Ví dụ họ gõ USD, bot trả về thông tin tỉ giá giữa đô la Mĩ và đồng Việt Nam. Cách này đơn giản nhất nhưng có điểm dở là người dùng sẽ gõ nhiều kiểu khác nhau và thậm chí gõ sai thì chúng ta xử lý rất mệt. Ví dụ: gõ usd hay usđ, ...

2. Hiển thị danh sách ngoại tệ kèm thêm con số, người dùng nhập số tương ứng thì bot xử lý. Cách này tiện hơn cách trên vì nhập số ít sai sót nhưng nó chỉ phù hợp với các nền tảng chat chỉ hỗ trợ chat chuỗi văn bản.

3. Hiển thị danh sách ngoại tệ dạng câu trả lời nhanh. Cách này khá đơn giản khi phát triển nhưng cũng khá tiện lợi cho người dùng nên tôi sẽ chọn nó cho chức năng này. Điểm dở của nó là danh sách trả lời nhanh này sẽ mất đi khi người dùng chọn câu trả lời nên không chọn lại được nữa. Tối đa hiển thị được 11 câu trả lời nhanh. Lưu ý trả lời nhanh không có chạy trên Messenger Lite.

4. Hiển thị ngoại tệ dạng danh sách cuộn. Cách này thường dùng để cung cấp các thông tin có nhiều nội dung và hình ảnh đi kèm. Mỗi lần hiển thị được 10 mục. Tôi thấy hiển thị danh sách của Đông Á dùng cách này khá ổn do nó có cả hình ảnh là lá cờ các quốc gia nữa. Tài liệu về danh sách cuộn ở đây (https://developers.facebook.com/docs/messenger-platform/send-messages/template/generic#carousel). Các bạn làm thử nhé!
*/
func processMessage(event *Messaging) {
	// Gửi hành động đã xem và đang trả lời
	sendAction(event.Sender, MarkSeen)
	sendAction(event.Sender, TypingOn)

	// Xử lý khi người dùng chọn trả lời nhanh
	if event.Message.QuickReply != nil {
		processQuickReply(event)
		return
	}

	// Xử lý khi người dùng gửi văn bản
	text := strings.ToLower(strings.TrimSpace(event.Message.Text))
	if text == RatePB {
		// Lưu nhóm ngoại tệ xem hiện tại
		exRateGroupMap[event.Sender.ID] = 1
		// Gửi danh sách ngoại tệ
		sendExchangeRateList(event.Sender)
	} else {
		// Gửi chuỗi nhận được sau khi chuyển sang chữ hoa
		sendText(event.Sender, strings.ToUpper(event.Message.Text))
	}

	// Gửi hành động đã trả lời xong
	sendAction(event.Sender, TypingOff)
}

// processPostBack - Khi nhấn vào "Bắt đầu" hoặc item menu
/**
Khi ấn vào nút "Bắt đầu" hoặc item của menu thì Facebook gửi Postback cho bot nên chúng ta phải xử lý nó trước đã. Các bạn đọc tài liệu về Postback ở đây (https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_postbacks)
*/
func processPostBack(event *Messaging) {
	// Gửi hành động đã xem và đang trả lời
	sendAction(event.Sender, MarkSeen)
	sendAction(event.Sender, TypingOn)

	switch event.PostBack.Payload {
	case GetStartedPB, RatePB:
		exRateGroupMap[event.Sender.ID] = 1
		sendExchangeRateList(event.Sender)
	}
	// Gửi hành động đã trả lời xong
	sendAction(event.Sender, TypingOff)
}

func getExchangeRateVCB() (*ExchangeRate, bool) {
	var exrate ExchangeRate

	req, err := http.NewRequest("GET", ExchangeRateVCBURL, nil)
	if err != nil {
		log.Errorf("getExchangeRateVCB: NewRequest: %s", err.Error())
		return &exrate, false
	}

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("getExchangeRateVCB: client.Do: %s", err.Error())
		return &exrate, false
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&exrate)
	if err != nil {
		log.Errorf("getExchangeRateVCB: xml.NewDecoder: %s", err.Error())
		return &exrate, false
	}

	return &exrate, true
}

func sendTextWithQuickReply(recipient *User, message string, replies []QuickReply) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Message: &ResMessage{
			Text:       message,
			QuickReply: replies,
		},
	}
	return sendFBRequest(FBMessageURL, &m)
}

func processQuickReply(event *Messaging) {
	recipient := event.Sender
	exRateGroup := exRateGroupMap[event.Sender.ID]

	switch event.Message.QuickReply.Payload {
	case "Next": // Trường hợp người dùng chọn "Xem tiếp"
		var i int
		// Kiểm tra nếu đã xem xong danh sách thì quay lại
		if exRateGroup*10 >= len(exRateList.Exrate) {
			exRateGroup = 1
		} else {
			exRateGroup++
		}
		exRateGroupMap[event.Sender.ID] = exRateGroup
		quickRep := []QuickReply{}
		// Mỗi lần hiển thị gồm 10 ngoại tệ
		for i = 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.Exrate); i++ {
			exrate := exRateList.Exrate[i]
			quickRep = append(quickRep, QuickReply{ContentType: "text", Title: exrate.CurrencyName, Payload: exrate.CurrencyCode})
		}
		// Thêm nút "Xem tiếp"
		quickRep = append(quickRep, QuickReply{ContentType: "text", Title: "Xem tiếp", Payload: "Next"})
		sendTextWithQuickReply(recipient, "GoBot cung cấp chức năng xem tỉ giá giữa các ngoại tệ và đồng Việt Nam.\nMời bạn chọn ngoại tệ:", quickRep)

	default: // Trường hợp người dùng chọn 1 nút trả lời nhanh
		var exRate Exrate
		// Kiểm tra coi payload nhận được khớp với item nào
		for i := 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.Exrate); i++ {
			if exRateList.Exrate[i].CurrencyCode == event.Message.QuickReply.Payload {
				exRate = exRateList.Exrate[i]
				break
			}
		}
		// Không tìm thấy item nào khớp
		if len(exRate.CurrencyCode) == 0 {
			sendText(recipient, "Không có thông tin về ngoại tệ này")
			return
		}
		// Trả về thông tin tìm được
		sendText(recipient, fmt.Sprintf("%s-VND\nGiá mua: %sđ\nGiá bán: %sđ\nGiá chuyển khoản: %sđ", exRate.CurrencyCode, exRate.Buy, exRate.Sell, exRate.Transfer))
	}
}

func sendExchangeRateList(recipient *User) {
	var (
		ok          bool
		i           int
		exRateGroup = exRateGroupMap[recipient.ID]
	)

	// Lấy danh sách ngoại tệ và lưu vào biến toàn cục exRateList
	exRateList, ok = getExchangeRateVCB()
	if !ok {
		sendText(recipient, "Có lỗi trong quá trình xử lý. Bạn vui lòng thử lại sau bằng cách gửi 'rate' cho tôi nhé. Cảm ơn!")
		return
	}

	quickRep := []QuickReply{}
	// Lấy nhóm 10 ngoại tệ
	for i = 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.Exrate); i++ {
		exrate := exRateList.Exrate[i]
		quickRep = append(quickRep, QuickReply{ContentType: "text", Title: exrate.CurrencyName, Payload: exrate.CurrencyCode})
	}

	quickRep = append(quickRep, QuickReply{ContentType: "text", Title: "Xem tiếp", Payload: "Next"})
	sendTextWithQuickReply(recipient, "GoBot cung cấp chức năng xem tỉ giá giữa các ngoại tệ và đồng Việt Nam.\nMời bạn chọn ngoại tệ:", quickRep)
}
