package instagram

import (
	"bytes"
)

func UrlUsers(userId string) string {
	return IgBaseURL.append("users/").append(userId).String()
}

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

func UrlRelationshipsFollows(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/follows").String()
}

func UrlRelationshipsFollowedBy(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/followed-by").String()
}

func UrlRelationshipsRequestedBy() string {
	return IgBaseURL.append("users/self/requested-by").String()
}

func UrlRelationshipsRelationship(userId string) string {
	return IgBaseURL.append("users/").append(userId).append("/relationship").String()
}

func UrlMedia(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).String()
}

func UrlMediaSearch() string {
	return IgBaseURL.append("media/search").String()
}

func UrlMediaPopular() string {
	return IgBaseURL.append("media/popular").String()
}

func UrlComments(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/comments").String()
}

func UrlCommentsDelete(mediaId, commentId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/comments/").append(commentId).String()
}

func UrlLikes(mediaId string) string {
	return IgBaseURL.append("media/").append(mediaId).append("/likes").String()
}

func UrlTags(tagName string) string {
	return IgBaseURL.append("tags/").append(tagName).String()
}

func UrlTagsMediaRecent(tagName string) string {
	return IgBaseURL.append("tags/").append(tagName).append("/media/recent").String()
}

func UrlTagsSearch() string {
	return IgBaseURL.append("tags/search").String()
}

func UrlLocations(locName string) string {
	return IgBaseURL.append("locations/").append(locName).String()
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

func (igu IgUrl) append(addendum string) IgUrl {
	var buffer bytes.Buffer
	buffer.WriteString(igu.String())
	buffer.WriteString(addendum)

	return IgUrl(buffer.String())
}

func (igu IgUrl) String() string {
	return string(igu)
}
