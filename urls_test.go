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
	x := UrlTags("selfie").String()
	y := IgBaseURL.append("tags/").append("selfie").String()

	if x != y {
		t.Errorf("Asked for %s, got %s", y, x)
	}
}

func TestUrlTagsSearch(t *testing.T) {
	x := UrlTagsSearch("selfies").String()
	y := IgBaseURL.append("tags/search?q=selfies").String()

	binaryChecker(x, y, t)
}

func TestUrlTagsMediaRecent(t *testing.T) {
	x := UrlTagsMediaRecent("boo", "1024", "512").String()
	y := IgBaseURL.append("tags/boo/media/recent?max_id=512&min_id=1024").String()

	binaryChecker(x, y, t)
}

func binaryChecker(x, y string, t *testing.T) {
	if x != y {
		t.Errorf("Asked for %s, got %s", y, x)
	}
}
