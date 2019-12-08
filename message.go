package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
)

const (
	// SubscriptionVerifyToken - chứa mã xác minh mà chúng ta đã khai báo khi thiết lập webhook
	SubscriptionVerifyToken = cfgSubscriptionVerifyToken

	// FBMessageURL - địa chỉ gửi phản hồi
	FBMessageURL = cfgFBMessageURL

	// FBMessengerProfileURL - địa chỉ để đăng ký Màn hình chào và Menu
	FBMessengerProfileURL = cfgFBMessengerProfileURL

	// PageToken       = "<Mã nhận được ở bước tạo mã khi tạo Facebook App>"
	PageToken = cfgPageToken

	// MessageResponse - Loại tin nhắn phản hồi có nhiều kiểu khác nhau, ở đây đơn giản tôi chọn là "RESPONSE", các loại khác các bạn tham khảo thêm ở đây (https://developers.facebook.com/docs/messenger-platform/send-messages#messaging_types)
	MessageResponse = "RESPONSE"

	// MarkSeen xác nhận bot đã xem tin, trên cửa số chat sẽ hiển thị avatar của trang ngay dưới câu của đối phương để báo việc đã xem
	MarkSeen = "mark_seen"
	// TypingOn hiển thị trạng thái báo bot đang gõ phản hồi
	TypingOn = "typing_on"
	// TypingOff ẩn trạng thái báo gõ phản hồi
	TypingOff = "typing_off"
)

type (
	// Khai báo omitempty nghĩa là nếu giá trị đó rỗng (0 với kiểu số, "" với chuỗi, [] với slice và nil với kiểu con trỏ) thì khi đóng JSON gửi đi, nó sẽ bị loại bỏ khỏi JSON.

	// Request Object
	Request struct {
		Object string `json:"object,omitempty"`
		Entry  []struct {
			ID        string      `json:"id,omitempty"`
			Time      int64       `json:"time,omitempty"`
			Messaging []Messaging `json:"messaging,omitempty"`
		} `json:"entry,omitempty"`
	}

	// Messaging Object
	Messaging struct {
		Sender    *User    `json:"sender,omitempty"`
		Recipient *User    `json:"recipient,omitempty"`
		Timestamp int      `json:"timestamp,omitempty"`
		Message   *Message `json:"message,omitempty"`

		PostBack *PostBack `json:"postback,omitempty"`
	}

	// PostBack - Khi ấn vào nút "Bắt đầu" hoặc item của menu thì Facebook gửi Postback cho bot nên chúng ta phải xử lý nó trước đã. Các bạn đọc tài liệu về Postback ở đây (https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_postbacks)
	PostBack struct {
		Title   string `json:"title,omitempty"`
		Payload string `json:"payload"`
	}

	// User Object
	User struct {
		ID string `json:"id,omitempty"`
	}

	// Message Object
	Message struct {
		MID  string `json:"mid,omitempty"`
		Text string `json:"text,omitempty"`

		// QuickReply - Trả lời nhanh
		QuickReply *QuickReply `json:"quick_reply,omitempty"`
	}

	// ResponseMessage Object
	ResponseMessage struct {
		MessageType string      `json:"messaging_type"`
		Recipient   *User       `json:"recipient"`
		Message     *ResMessage `json:"message,omitempty"`

		// Phản hồi hành động chat. Thuộc tính này dạng chuỗi có 3 giá trị:
		// - "typing_on": hiển thị trạng thái báo bot đang gõ phản hồi.
		// - "typing_off": ẩn trạng thái báo gõ phản hồi.
		// - "mark_seen": xác nhận bot đã xem tin, trên cửa số chat sẽ hiển thị avatar của trang ngay dưới câu của đối phương để báo việc đã xem.
		Action string `json:"sender_action,omitempty"`
	}

	// ResMessage Object
	ResMessage struct {
		Text       string       `json:"text,omitempty"`
		QuickReply []QuickReply `json:"quick_replies,omitempty"`
	}

	// QuickReply - Trả lời nhanh, xem tại đây (https://developers.facebook.com/docs/messenger-platform/send-messages/quick-replies#text)
	QuickReply struct {
		ContentType string `json:"content_type,omitempty"`
		Title       string `json:"title,omitempty"`
		Payload     string `json:"payload"`
	}
)

// Xác minh webhook
/**
Quy trình xác minh webhook của Facebook app như sau:

1. Nền tảng Messenger của Facebook sẽ gửi yêu cầu GET tới webhook mà chúng ta khai báo với 3 tham số:
- hub.mode: luôn là "subscribe"
- hub.verify_token: chứa mã xác minh mà chúng ta đã khai báo khi thiết lập webhook. Chuỗi này tôi đã chọn là "GoBot". Bạn có thể chọn tùy thích.
- hub.challenge: chuỗi số mà chúng ta phải gửi lại khi xác nhận kết quả xác minh.

2. Chúng ta xác minh rằng mã Facebook gửi đến ở hub.verify_token khớp với mã mình chọn và phản hồi bằng tham số hub.challenge.

3. Nền tảng Messenger đăng ký webhook của chúng ta với ứng dụng đã tạo.
*/
func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && token == SubscriptionVerifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Error, wrong validation token"))
	}
}

// Xử lý sự kiện
/**
Cái chúng ta cần quan tâm lúc này là facebook sẽ gửi cho chúng ta cái gì. Trên trang tài liệu Facebook (https://developers.facebook.com/docs/messenger-platform/webhook#format) cho chúng ta biết đó là 1 định dạng JSON có cấu trúc như sau:

{
    "object":"page",
    "entry":[
        {
            "id":"<PAGE_ID>",
            "time":1458692752478,
            "messaging":[
                {
                    "sender":{
                        "id":"<PSID>"
                    },
                    "recipient":{
                        "id":"<PAGE_ID>"
                    },
					"timestamp": 1574777347965,

					// Khi nhấn vào "Bắt đầu" hoặc item menu
                    "postback": {
                        "title": "Tỉ giá hối đoái",
                        "payload": "rate"
					}

					// Khi người dùng gõ nội dung bất kỳ vào khung chat
					"message": {
                        "mid": "<MID>",
                        "text": "..."
                    }
                    ...
                }
            ]
        }
    ]
}

Tuy nhiên nếu chưa xử lý gì mà chờ xem rồi mới xử lý thì cần lưu ý những điểm sau:
- Dù không xử lý gì thì ít nhất luôn phản hồi 200 OK về cho Facebook nếu không muốn bị xếp vào diện lỗi.
- Nếu quá 20 giây từ khi gửi mà không thấy bot server phản hồi, Facebook sẽ gửi lặp lại tin đó nhiều lần nữa.
- Sau 15 phút mà server không phản hồi thì Facebook xem như webhook lỗi và cảnh báo.
- Sau 8 giờ cảnh báo thì webhook bạn đăng ký sẽ bị vô hiệu hóa và nếu muốn bạn phải xác minh lại.
*/
func processWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Message not supported"))
		return
	}

	if req.Object == "page" {
		for _, entry := range req.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil {
					processMessage(&event)
				} else if event.PostBack != nil {
					processPostBack(&event)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Got your message"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Message not supported"))
	}
}

// Gửi tin phản hồi
/**
Facebook có quy định cấu trúc JSON phản hồi và URL nhận tin nhắn tại đây (https://developers.facebook.com/docs/messenger-platform/send-messages#send_api_basics).
Cấu trúc này như sau:
{
	"messaging_type": "<MESSAGING_TYPE>",
	"recipient":{
  		"id":"<PSID>"
	},
	"message":{
  		"text":"hello, world!"
	}
}
*/
func sendFBRequest(url string, m interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&m)
	if err != nil {
		log.Error("sendFBRequest:json.NewEncoder: " + err.Error())
		return err
	}

	//
	fmt.Println("body to send", body)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error("sendFBRequest:http.NewRequest:" + err.Error())
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = "access_token=" + PageToken
	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("sendFBRequest:client.Do: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	return nil
}

func sendText(recipient *User, message string) error {
	return sendTextWithQuickReply(recipient, message, nil)
}

func sendAction(recipient *User, action string) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Action:      action,
	}
	return sendFBRequest(FBMessageURL, &m)
}
