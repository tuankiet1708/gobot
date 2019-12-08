package main

import "github.com/apex/log"

// Màn hình chào và Menu
/**
Màn hình chào (https://developers.facebook.com/docs/messenger-platform/discovery/welcome-screen) là màn hình hiển thị khi một người dùng mở cửa sổ chat với page lần đầu tiên hoặc ngay sau khi họ chọn xóa dữ liệu chat với page. Màn hình này có 1 lời chào và một nút "Bắt đầu" hoặc "Get Started" để giúp người dùng chọn tránh bỡ ngỡ ban đầu. Khi họ bấm nút, bot sẽ được 1 sự kiện postback

Menu (https://developers.facebook.com/docs/messenger-platform/send-messages/persistent-menu) là danh sách các chức năng luôn hiển thị để người dùng có thể chọn nhanh chức năng họ cần. Khi họ chọn, bot sẽ nhận được sự kiện postback tương ứng.

Dựa trên tài liệu Facebook liên quan ở trên, tôi khai báo các struct và tạo hàm thiết lập màn hình chào và menu như sau:
*/
type (
	PageProfile struct {
		Greeting       []Greeting       `json:"greeting,omitempty"`
		GetStarted     *GetStarted      `json:"get_started,omitempty"`
		PersistentMenu []PersistentMenu `json:"persistent_menu,omitempty"`
	}

	Greeting struct {
		Locale string `json:"locale,omitempty"`
		Text   string `json:"text,omitempty"`
	}

	GetStarted struct {
		Payload string `json:"payload,omitempty"`
	}

	PersistentMenu struct {
		Locale   string `json:"locale"`
		Composer bool   `json:"composer_input_disabled"`
		CTAs     []CTA  `json:"call_to_actions"`
	}

	CTA struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		URL     string `json:"url,,omitempty"`
		Payload string `json:"payload"`
		CTAs    []CTA  `json:"call_to_actions,omitempty"`
	}
)

const (
	GetStartedPB = "GetStarted"
	RatePB       = "rate"
)

/**
Hàm registerGreetingnMenu khi nào cần thay đổi chúng ta mới gọi bởi vì mỗi khi gọi, menu sẽ không có hiệu lực ngay mà phải refresh mới thấy. Do đó tôi thường để nó đầu hàm main, sau đó return luôn để chạy mỗi nó. Chạy xong thì comment 2 dòng này để hàm main hoạt động bình thường lại như trước.
*/
func registerGreetingnMenu() bool {
	profile := PageProfile{
		Greeting: []Greeting{
			{
				Locale: "default",
				Text:   "ChatbotByLeo - Ứng dụng cung cấp thông tin tỉ giá hối đoái",
			},
		},
		GetStarted: &GetStarted{Payload: GetStartedPB},
		PersistentMenu: []PersistentMenu{
			{
				Locale:   "default",
				Composer: false,
				CTAs: []CTA{
					{
						Type:    "postback",
						Title:   "Tỉ giá hối đoái",
						Payload: RatePB,
					},
				},
			},
		},
	}
	err := sendFBRequest(FBMessengerProfileURL, &profile)
	if err != nil {
		log.Error("registerGreetingnMenu:" + err.Error())
		return false
	}
	return true
}
