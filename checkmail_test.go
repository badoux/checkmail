package checkmail

import (
	"fmt"
	"os"
	"testing"

)

var (
	samples = []struct {
		mail    string
		format  bool
		account bool //host+user
	}{
		{mail: "florian@carrere.cc", format: true, account: true},
		{mail: "support@g2mail.com", format: true, account: false},
		{mail: " florian@carrere.cc", format: false, account: false},
		{mail: "florian@carrere.cc ", format: false, account: false},
		{mail: "test@912-wrong-domain902.com", format: true, account: false},
		{mail: "0932910-qsdcqozuioqkdmqpeidj8793@gmail.com", format: true, account: false},
		{mail: "@gmail.com", format: false, account: false},
		{mail: "test@gmail@gmail.com", format: false, account: false},
		{mail: "test test@gmail.com", format: false, account: false},
		{mail: " test@gmail.com", format: false, account: false},
		{mail: "test@wrong domain.com", format: false, account: false},
		{mail: "é&ààà@gmail.com", format: false, account: false},
		{mail: "admin@busyboo.com", format: true, account: false},
		{mail: "a@gmail.fi", format: true, account: false},
	}
)

func TestValidateHost(t *testing.T) {
	for _, s := range samples {
		if !s.format {
			continue
		}

		err := ValidateHost(s.mail)
		if err != nil && s.account == true {
			t.Errorf(`"%s" => unexpected error: "%v"`, s.mail, err)
		}
		if err == nil && s.account == false {
			t.Errorf(`"%s" => expected error`, s.mail)
		}
	}
}

func TestValidateHostAndUser(t *testing.T) {
	var (
		serverHostName = getenv(t, "self_hostname")
		serverMailAddress = getenv(t, "self_mail")
	)
	for _, s := range samples {
		if !s.format {
			continue
		}

		err := ValidateHostAndUser(serverHostName, serverMailAddress, s.mail)
		if err != nil && s.account == true {
			t.Errorf(`"%s" => unexpected error: "%v"`, s.mail, err)
		}
		if err == nil && s.account == false {
			t.Errorf(`"%s" => expected error`, s.mail)
		}
	}
}

func TestValidateFormat(t *testing.T) {
	for _, s := range samples {
		err := ValidateFormat(s.mail)
		if err != nil && s.format == true {
			t.Errorf(`"%s" => unexpected error: "%v"`, s.mail, err)
		}
		if err == nil && s.format == false {
			t.Errorf(`"%s" => expected error`, s.mail)
		}
	}
}

func getenv(t *testing.T, name string) (value string) {
	name = "test_checkmail_"+name
	if value = os.Getenv(name); value =="" {
		panic(fmt.Errorf("enviroment variable %q is not defined", name))
	}
	return
}