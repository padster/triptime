package fb

type RequestBody struct {
    Entry       []Entry         `json:"entry"`
}

type Entry struct {
    Message     []Message       `json:"messaging"`
}

type Message struct {
    Sender      User            `json:"sender"`
    Recipient   User            `json:"recipient"`
    Message     MessageData     `json:"message"`
    Delivery    *Delivery       `json:"delivery,omitempty"`
    Postback    *Postback       `json:"postback,omitempty"`
}

type Delivery struct {
    MessageIds  []string        `json:"mids"`
    Watermark   int64           `json:"watermark"`
    Seq         int             `json:"seq"`
}

type Postback struct {
    Payload     string          `json:"payload"`
}

type User struct {
    Id          string          `json:"id"`
}

type MessageData struct {
    Mid         string          `json:"mid,omitempty"`
    Text        string          `json:"text,omitempty"`
    Attachment  []Attachment    `json:"attachments,omitempty"`
}

type Attachment struct {
    Title       string          `json:"title"`
    Type        string          `json:"type"`
    Payload     Payload         `json:"payload"`
}

type Payload struct {
    Coordinates Coordinates     `json:"coordinates"` 
}

type Coordinates struct {
    Lat         float64         `json:"lat"`
    Long        float64         `json:"long"`
}

type OutboundMessage struct {
    Recipient   User            `json:"recipient"`
    Message     OutMessageData  `json:"message"`
}

type Request struct {
    Recipient User              `json:"recipient"`
    Message   OutMessageData    `json:"message"`
}

type OutMessageData struct {
    Text        string          `json:"text,omitempty"`
    Attachment  *OutAttachment  `json:"attachment,omitempty"`
}

type OutAttachment struct {
    Type    string            `json:"type"`
    Payload AttachmentPayload `json:"payload"`
}

type AttachmentPayload interface{}

type Button struct {
    Type    string `json:"type"`
    Title   string `json:"title,omitempty"`
    URL     string `json:"url,omitempty"`
    Payload string `json:"payload,omitempty"`
}

type ButtonPayload struct {
    Type    string `json:"template_type"`
    Text    string   `json:"text,omitempty"`
    Buttons []Button `json:"buttons,omitempty"`
}

func (bt *ButtonPayload) AddButton(btn Button) {
    bt.Buttons = append(bt.Buttons, btn) 
} 
