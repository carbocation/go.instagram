package instagram

import "testing"

func TestAppend(t *testing.T) {
	u := IgUrl("http://www.google.com/")
	u = u.append("GOOG")

	if u.String() != "http://www.google.com/GOOG" {
		t.Errorf("Asked for %s, got %s", "http://www.google.com/GOOG", u.String())
	}
}

func TestUrlTags(t *testing.T) {
	x := UrlTags("selfie")
	y := IgBaseURL.append("tags/").append("selfie").String()

	if x != y {
		t.Errorf("Asked for %s, got %s", y, x)
	}
}
