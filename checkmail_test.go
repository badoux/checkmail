package checkmail

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

var (
	samples = []struct {
		mail    string
		format  bool
		mx      bool
		account bool //host+user
	}{
		{mail: "florian@carrere.cc", format: true, mx: true, account: true},
		{mail: "support@g2mail.com", format: true, mx: true, account: false},
		{mail: " florian@carrere.cc", format: false, mx: false, account: false},
		{mail: "florian@carrere.cc ", format: false, mx: false, account: false},
		{mail: "test@912-wrong-domain902.com", format: true, mx: false, account: false},
		{mail: "0932910-qsdcqozuioqkdmqpeidj8793@gmail.com", format: true, mx: true, account: false},
		{mail: "@gmail.com", format: false, mx: false, account: false},
		{mail: "test@gmail@gmail.com", format: false, mx: false, account: false},
		{mail: "test test@gmail.com", format: false, mx: false, account: false},
		{mail: " test@gmail.com", format: false, mx: false, account: false},
		{mail: "test@wrong domain.com", format: false, mx: false, account: false},
		{mail: "é&ààà@gmail.com", format: false, mx: false, account: false},
		{mail: "admin@notarealdomain12345.com", format: true, mx: false, account: false},
		{mail: "a@gmail.xyz", format: true, mx: false, account: false},
		{mail: "", format: false, mx: false, account: false},
		{mail: "not-a-valid-email", format: false, mx: false, account: false},
	}
)

func TestValidateHost(t *testing.T) {
	for _, s := range samples {
		if !s.format {
			continue
		}

		err := ValidateHost(s.mail)
		if err != nil && s.account == true {
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") {
				t.Logf(`"%s" => live server unreachable: "%v"`, s.mail, err)
			} else {
				t.Errorf(`"%s" => unexpected error: "%v"`, s.mail, err)
			}
		}
		if err == nil && s.account == false {
			t.Errorf(`"%s" => expected error`, s.mail)
		}
	}
}

func TestValidateMX(t *testing.T) {
	for _, s := range samples {
		if !s.format {
			continue
		}

		err := ValidateMX(s.mail)
		if err != nil && s.mx == true {
			t.Errorf(`"%s" => unexpected error: "%v"`, s.mail, err)
		}
		if err == nil && s.mx == false {
			t.Errorf(`"%s" => expected error`, s.mail)
		}
	}
}

func TestValidateHostAndUser(t *testing.T) {
	var (
		serverHostName    = getenv(t, "self_hostname")
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

func TestEmptyMXRecord(t *testing.T) {
	origLookup := lookupMX
	defer func() { lookupMX = origLookup }()

	testEmail := "user@example.com"

	// Empty slice returned from lookup (Issue #43 reproduction)
	lookupMX = func(host string) ([]*net.MX, error) {
		return []*net.MX{}, nil
	}

	if err := ValidateMX(testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateMX with empty MX slice = %v; want ErrNoSuchHost", err)
	}
	if err := ValidateHost(testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateHost with empty MX slice = %v; want ErrNoSuchHost", err)
	}
	if err := ValidateHostAndUser("client.example.com", "sender@client.example.com", testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateHostAndUser with empty MX slice = %v; want ErrNoSuchHost", err)
	}

	// Nil slice returned from lookup
	lookupMX = func(host string) ([]*net.MX, error) {
		return nil, nil
	}

	if err := ValidateMX(testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateMX with nil MX slice = %v; want ErrNoSuchHost", err)
	}
	if err := ValidateHost(testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateHost with nil MX slice = %v; want ErrNoSuchHost", err)
	}
	if err := ValidateHostAndUser("client.example.com", "sender@client.example.com", testEmail); err != ErrNoSuchHost {
		t.Errorf("ValidateHostAndUser with nil MX slice = %v; want ErrNoSuchHost", err)
	}

	// DNS error returned from lookup
	lookupMX = func(host string) ([]*net.MX, error) {
		return nil, errors.New("dns lookup failed")
	}

	if err := ValidateMX(testEmail); err != ErrUnresolvableHost {
		t.Errorf("ValidateMX with DNS error = %v; want ErrUnresolvableHost", err)
	}
	if err := ValidateHost(testEmail); err != ErrUnresolvableHost {
		t.Errorf("ValidateHost with DNS error = %v; want ErrUnresolvableHost", err)
	}
	if err := ValidateHostAndUser("client.example.com", "sender@client.example.com", testEmail); err != ErrUnresolvableHost {
		t.Errorf("ValidateHostAndUser with DNS error = %v; want ErrUnresolvableHost", err)
	}
}

func getenv(t *testing.T, name string) (value string) {
	name = "test_checkmail_" + name
	if value = os.Getenv(name); value == "" {
		t.Skip(fmt.Sprintf("enviroment variable %q is not defined", name))
	}
	return
}
