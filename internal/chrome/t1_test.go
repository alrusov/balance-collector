package chrome

import (
	"testing"
)

//----------------------------------------------------------------------------------------------------------------------------//

func TestParse(t *testing.T) {
	tasks := []string{
		"",
		"Navigate https://my.onlime.ru/session/logout ",
		"Navigate(https://my.onlime.ru/session/logout)",
		"WaitVisible(form[id=\"login_form\"])",

		"  SendKeys  (  form[id=\"login_form\"]\\,\\, \\, input[name=\"login_credentials[login]\"]  ,   $Login  )  ",
		"SendKeys(form[id=\"login_form\"] input[name=\"login_credentials[password]\"], $Password)",
		"Click(form[id=\"login_form\"] input[id=\"login_button\"])",

		"Float(div[id=\"account_info_block\"] > p:nth-child(2) > big, Баланс)",
		"Float(div[id=\"account_info_block\"] > p:nth-child(3) > big, Бонусы)",
		"Float(div[id=\"account_info_block\"] > p:nth-child(4) > big, Дней до блокировки)",
		"String(,Тариф)",
		"String(div[id=\"account_info_block\"] > p:nth-child(6) > strong, Статус)",
	}

	// смотрим в дебагере
	_, err := New(tasks)
	if err != nil {
		//t.Error(err)
	}
}

//----------------------------------------------------------------------------------------------------------------------------//
