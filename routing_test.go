package routing

import (
	"regexp"
	"testing"
)

func TestGetDefaultRouteLinux(t *testing.T) {
	result, err := FindLinuxDefaultGW()
	if err != nil {
		t.Errorf("Calling routing library failed %s %s", result, err.Error())
	}
	expected, regexpErr := regexp.MatchString("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$", result)
	if regexpErr != nil {
		t.Errorf("Calling regexp match inside test failed %t %s", expected, regexpErr.Error())
	}

	if !expected {
		t.Errorf("Did not match IP address %s %s", result, "xxx.xxx.xxx.xxx")
	}
}

func TestGetLinuxRoutingTable(t *testing.T) {
	result, err := GetLinuxRoutingTable()
	if err != nil {
		t.Errorf("Calling routing library failed %s %s", "", err.Error())
	}

	expected, regexpErr := regexp.MatchString("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$", result[0].Gateway)
	if regexpErr != nil {
		t.Errorf("Calling regexp match inside test failed %t %s", expected, regexpErr.Error())
	}

	if !expected {
		t.Errorf("Did not match IP address %s %s", result[0].Gateway, "xxx.xxx.xxx.xxx")
	}
}
