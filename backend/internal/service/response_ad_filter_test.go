package service

import "testing"

func TestSanitizeModelAdText_RemovesQQAdLine(t *testing.T) {
	input := "正常回答第一段\n服务暂停，公益QQ群通知：12454546565.\n正常回答第二段"
	got := sanitizeModelAdText(input)
	want := "正常回答第一段\n正常回答第二段"
	if got != want {
		t.Fatalf("unexpected sanitized text\nwant: %q\ngot:  %q", want, got)
	}
}

func TestSanitizeModelAdResponseBody_OpenAIChatChoice(t *testing.T) {
	body := []byte(`{"id":"chatcmpl_1","choices":[{"index":0,"message":{"role":"assistant","content":"可正常使用。\n服务暂停，公益QQ群通知：12454546565."}}]}`)
	got := sanitizeModelAdResponseBody(body)
	want := `{"id":"chatcmpl_1","choices":[{"index":0,"message":{"role":"assistant","content":"可正常使用。"}}]}`
	if string(got) != want {
		t.Fatalf("unexpected body\nwant: %s\ngot:  %s", want, string(got))
	}
}

func TestSanitizeModelAdResponseBody_OpenAIResponsesOutputText(t *testing.T) {
	body := []byte(`{"id":"resp_1","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"答复内容\n服务暂停，公益QQ群通知：12454546565."}]}]}`)
	got := sanitizeModelAdResponseBody(body)
	want := `{"id":"resp_1","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"答复内容"}]}]}`
	if string(got) != want {
		t.Fatalf("unexpected body\nwant: %s\ngot:  %s", want, string(got))
	}
}

func TestSanitizeModelAdResponseBody_NoChange(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"content":"正常回答，没有广告"}}]}`)
	got := sanitizeModelAdResponseBody(body)
	if string(got) != string(body) {
		t.Fatalf("expected body to remain unchanged\ngot: %s", string(got))
	}
}
