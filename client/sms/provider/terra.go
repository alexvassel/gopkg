package provider

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/severgroup-tt/gopkg-app/app"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/severgroup-tt/gopkg-errors"
	"github.com/severgroup-tt/gopkg-logger"
)

// https://terasms.ru/documentation/api/http/errors
var terraErrorsText = map[string]string{
	"-1":   "Неверный логин или пароль",
	"-20":  "Пустой текст сообщения",
	"-30":  "Пустой номер абонента",
	"-40":  "Неправильно задан номер абонента",
	"-45":  "Превышено количество номеров",
	"-50":  "Неправильно задано имя отправителя",
	"-60":  "Рассылка по данному направлению недоступна",
	"-70":  "Недостаточно средств на счете",
	"-80":  "Не установлена стоимость рассылки по данному направлению",
	"-90":  "Рассылка запрещена",
	"-100": "Не указаны необходимые параметры",
	"-110": "Номер в черном списке",
	"-120": "Некорректно задано время отложенной отправки",
	"-130": "Некорректно задано временное окно отправки",
	"-140": "Передан некорректный ID рассылки",
	"-160": "Превышен дневной лимит рассылки (Вы можете установить максимальную сумму ежедневной рассылки после согласования с Вашим менеджером)",
}

type terraProvider struct {
	URL       string
	Login     string
	Sender    string
	Password  string
	showInfo  bool
	showError bool
}

type terraSendRequest struct {
	Login      string         `json:"login"`
	SmsPackage []terraMessage `json:"smsPackage"`
	Sign       string         `json:"sign"`
}

type terraSendStatusResponse struct {
	Status      *int    `json:"status"`
	Description *string `json:"status_description"`
}

type terraSendMessageResponse struct {
	SmsID     int    `json:"sms_id"`
	MessageID string `json:"message_id"`
}

type terraMessage struct {
	ID      int    `json:"sms_id"`
	Phone   int64  `json:"target"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

func NewTerraProvider(url, login, sender, password string) IProvider {
	return &terraProvider{
		URL:      url,
		Login:    login,
		Sender:   sender,
		Password: password,
	}
}

func (c terraProvider) Connect(showInfo, showError bool) (ISender, app.PublicCloserFn, error) {
	c.showInfo = showInfo
	c.showError = showError
	return c, nil, nil
}

func (c terraProvider) Send(ctx context.Context, phone int64, message string) error {
	messages := []terraMessage{{
		ID:      1,
		Sender:  c.Sender,
		Phone:   phone,
		Message: message,
	}}

	request := terraSendRequest{
		Login:      c.Login,
		SmsPackage: messages,
		Sign:       c.getSign(messages),
	}

	bts, _ := json.Marshal(&request)
	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(bts))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// is status response?
	var statusResponse terraSendStatusResponse
	if err = json.Unmarshal(body, &statusResponse); err == nil {
		if statusResponse.Status != nil && *statusResponse.Status < 0 {
			return errors.Internal.Err(ctx, "SMS error: "+*statusResponse.Description)
		}
		return nil
	}

	// is message response?
	var msgResponse []terraSendMessageResponse
	err = json.Unmarshal(body, &msgResponse)
	if err == nil {
		for _, s := range msgResponse {
			if s.MessageID == "" || strings.Contains(s.MessageID, "-") {
				if c.showError {
					logger.Error(ctx, "Can't send messages: request = %+v, error = %s", request, c.getErrorText(s.MessageID))
				}
				return errors.Internal.Err(ctx, fmt.Sprintf("Couldn't sent SMS to %v", phone))
			} else if c.showInfo {
				logger.Info(ctx, fmt.Sprintf("Send messages - terraSmsId: %v, msgId: %v, phone: %v", s.SmsID, s.MessageID, phone))
			}
		}
		return nil
	}

	return errors.Internal.Err(ctx, fmt.Sprintf("Can't parse response from terra sms: %v, error: %v", body, err))
}

func (c terraProvider) getErrorText(code string) string {
	if err, ok := terraErrorsText[code]; ok {
		return err
	}
	return "Неопознанная ошибка: " + code
}

func (c terraProvider) getSign(messages []terraMessage) string {
	var sign string
	sign = "login=" + c.Login

	for _, v := range messages {
		sign += "message=" + v.Message + "sender=" + c.Sender + "sms_id=" + strconv.Itoa(v.ID) + "target=" + strconv.FormatInt(v.Phone, 10)
	}

	sign += c.Password
	return fmt.Sprintf("%x", md5.Sum([]byte(sign)))
}
