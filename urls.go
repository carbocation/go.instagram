package instagram

import (
	"bytes"
	"net/url"
)

type IgUrl string //URLs that we have constructed to be known-good instagram URLs

//Add a query parameter while preserving current query parameters (overwriting any dupes)
func (igu IgUrl) AddQuery(key, value string) IgUrl {
	//Exit early & graciously if they haven't provided this parameter
	if value == "" {
		return igu
	}

	u, _ := url.Parse(igu.String())
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()

	return IgUrl(u.String())
}

/*
func (igu IgUrl) AddToken() IgUrl {
	u, _ := url.Parse(igu.String())
	q := u.Query()
	q.Set("access_token", )
	u.RawQuery = q.Encode()

	return IgUrl(u.String())
}
*/

func (igu IgUrl) String() string {
	return string(igu)
}

func (igu IgUrl) append(addendum string) IgUrl {
	var buffer bytes.Buffer
	buffer.WriteString(igu.String())
	buffer.WriteString(addendum)

	return IgUrl(buffer.String())
}

func UrlUsers(userId string) string {
	return IgBaseURL.append("users/").append(userId).String()
}

func UrlRelationshipsFollows(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/follows").String()
}

func UrlRelationshipsFollowedBy(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/followed-by").String()
}

func UrlMedia(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).String()
}

func UrlMediaPopular() string {
	return IgBaseURL.append("media/popular").String()
}

func UrlComments(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/comments").String()
}

func UrlLikes(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/likes").String()
}

func UrlTags(tagName string) string {
	return IgBaseURL.append("tags/").append(tagName).String()
}

func UrlTagsMediaRecent(tagName, minId, maxId string) string {
	return IgBaseURL.append("tags/").append(tagName).append("/media/recent").AddQuery("min_id", minId).AddQuery("max_id", maxId).String()
}

func UrlTagsSearch(Q string) string {
	return IgBaseURL.append("tags/search").AddQuery("q", Q).String()
}

func UrlLocations(locName string) string {
	return IgBaseURL.append("locations/").append(locName).String()
}

/*
Incomplete
TODO: Complete
*/

func UrlUsersSelfFeed() string {
	return IgBaseURL.append("users/self/feed").String()
}

func UrlUsersMediaRecent(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/media/recent").String()
}

func UrlUsersSelfMediaLiked() string {
	return IgBaseURL.append("users/self/media/liked").String()
}

func UrlUsersSearch() string {
	return IgBaseURL.append("users/").String()
}

func UrlRelationshipsRequestedBy() string {
	return IgBaseURL.append("users/self/requested-by").String()
}

func UrlRelationshipsRelationship(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/relationship").String()
}

func UrlMediaSearch() string {
	return IgBaseURL.append("media/search").String()
}

func UrlCommentsDelete(mediaId, commentId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/comments/").append(commentId).String()
}

func UrlLocationsMediaRecent(locName string) string {
	return IgBaseURL.append("locations/").append(locName).append("/media/recent").String()
}

func UrlLocationsSearch() string {
	return IgBaseURL.append("locations/search").String()
}

func UrlGeographiesMediaRecent(geoId string) string {
	return IgBaseURL.append("geographies/").append(geoId).append("/media/recent").String()
}
