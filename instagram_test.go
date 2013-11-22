package instagram

import (
	//"fmt"
	"testing"
)

func TestQueryPublic(t *testing.T) {
	url := UrlTags("selfie")
	ig, err := NewInstagram("abcdefg")
	if err != nil {
		t.Error(err)
	}

	_, err = ig.QueryPublic(url)
	if err != nil {
		t.Error(err)
	}

	//fmt.Printf("%+v, %d", response, len(response.Data))
}
