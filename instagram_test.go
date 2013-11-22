package instagram

import (
	//"fmt"
	"testing"
)

func TestQueryPublic(t *testing.T) {
	Initialize(&Cfg{ClientID: "abcdefg"})

	url := UrlTags("selfie")

	_, err := QueryPublic(url)
	if err != nil {
		t.Error(err)
	}

	//fmt.Printf("%+v, %d", response, len(response.Data))
}
