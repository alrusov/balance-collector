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

func TestFloat(t *testing.T) {
	type paramsBlock struct {
		src    string
		expect float64
		isErr  bool
	}

	params := []paramsBlock{
		{`0.666`, 0.666, false},
		{`.666`, 0.666, false},
		{`6`, 6., false},
		{`6.`, 6., false},
		{`6.0`, 6., false},
		{`6.1`, 6.1, false},
		{`16`, 16., false},
		{`16.`, 16., false},
		{`16.0`, 16., false},
		{`16.1`, 16.1, false},
		{`qwerty:    26,0  `, 26., false},
		{`   26,1   `, 26.1, false},
		{`qwerty:    -26,0  `, -26., false},
		{`   −   26,1   `, -26.1, false},
		{`   &minus;   26,2   `, -26.2, false},
	}

	for i, p := range params {
		i++

		v, err := Float(p.src)
		if p.isErr && err == nil {
			t.Errorf(`%d: no error, error expected`, i)
		} else if !p.isErr && err != nil {
			t.Errorf(`%d: error: %s`, i, err.Error())
		} else if v != p.expect {
			t.Errorf(`%d: got %#v, %#v expected`, i, v, p.expect)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//
