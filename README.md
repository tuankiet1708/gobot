# Facebook Messenger Chatbot with Golang

> Chi tiết xem [tại đây](https://laptrinhgo.blogspot.com/2018/10/bai-37-gioi-thieu-facebook-messenger.html)

- Server chatbot phải xử lý 2 request GET và POST từ Facebook để xác minh server và xử lý tin nhắn từ người dùng trên page.
- Để Facebook có thể truy cập đến server trên máy chúng ta thì ngrok là một lựa chọn phù hợp. Ngoài chuyện tạo URL hỗ trợ https, ngrok còn cung cấp thông tin gửi nhận để chúng ta tiện theo dõi khi lập trình.
- Những gì Facebook gửi chúng ta phải xử lý và phản hồi trong vòng 20 giây nếu không Facebook sẽ gửi lại và nếu kéo dài dẫn đến webhook bị vô hiệu hóa.
- Bot cung cấp chức năng thông tin tỉ giá hối đoái, lấy dữ liệu từ VCB dạng XML.
- Trả lời nhanh là công cụ nền tảng Messenger hỗ trợ giúp tạo các nút bên dưới nội dung chat để người dùng có thể chọn cho câu trả lời của họ mà không phải gõ. Lưu ý: Các lựa chọn mất sau khi người dùng chọn, tối đa 11 lựa chọn và không hiển thị được ở Messenger Lite.
- Màn hình chào hiển thị khi người dùng lần đầu chat với page và có nút "Bắt đầu" để kích hoạt tương tác giữa người dùng và chatbot.
- Menu là công cụ luôn hiển thị để người dùng chọn khi cần đến nhanh một chức năng nào đó.
- Khi người dùng ấn chọn nút "Bắt đầu" hoặc menu, bot sẽ nhận được sự kiện postback.