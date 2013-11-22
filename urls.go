package instagram

import (
	"bytes"
	"net/url"
)

type IgUrl string //URLs that we have constructed to be known-good instagram URLs with helper methods

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

func (igu IgUrl) String() string {
	return string(igu)
}

func (igu IgUrl) addClientId(clientId string) IgUrl {
	return igu.AddQuery("client_id", clientId)
}

func (igu IgUrl) append(addendum string) IgUrl {
	var buffer bytes.Buffer
	buffer.WriteString(igu.String())
	buffer.WriteString(addendum)

	return IgUrl(buffer.String())
}

/*
Fully-specified URL query functions providing all options that the Instagram API provides:
*/

func UrlUsers(userId string) IgUrl {
	return IgBaseURL.append("users/").append(userId)
}

func UrlRelationshipsFollows(userId string) IgUrl {
	return IgBaseURL.append("users/").append(userId).append("/follows")
}

func UrlRelationshipsFollowedBy(userId string) IgUrl {
	return IgBaseURL.append("users/").append(userId).append("/followed-by")
}

func UrlMedia(mediaId string) IgUrl {
	return IgBaseURL.append("media/").append(mediaId)
}

func UrlMediaPopular() IgUrl {
	return IgBaseURL.append("media/popular")
}

func UrlComments(mediaId string) IgUrl {
	return IgBaseURL.append("media/").append(mediaId).append("/comments")
}

func UrlLikes(mediaId string) IgUrl {
	return IgBaseURL.append("media/").append(mediaId).append("/likes")
}

func UrlTags(tagName string) IgUrl {
	return IgBaseURL.append("tags/").append(tagName)
}

func UrlTagsMediaRecent(tagName, minId, maxId string) IgUrl {
	return IgBaseURL.append("tags/").append(tagName).append("/media/recent").AddQuery("min_id", minId).AddQuery("max_id", maxId)
}

func UrlTagsSearch(Q string) IgUrl {
	return IgBaseURL.append("tags/search").AddQuery("q", Q)
}

func UrlLocations(locName string) IgUrl {
	return IgBaseURL.append("locations/").append(locName)
}

/*
Incomplete URL query functions
These do NOT yet provide all options that the Instagram API provides
TODO: Complete them
*/

func UrlUsersSelfFeed() IgUrl {
	return IgBaseURL.append("users/self/feed")
}

func UrlUsersMediaRecent(userId string) IgUrl {
	return IgBaseURL.append("users/").append(userId).append("/media/recent")
}

func UrlUsersSelfMediaLiked() IgUrl {
	return IgBaseURL.append("users/self/media/liked")
}

func UrlUsersSearch() IgUrl {
	return IgBaseURL.append("users/")
}

func UrlRelationshipsRequestedBy() IgUrl {
	return IgBaseURL.append("users/self/requested-by")
}

func UrlRelationshipsRelationship(userId string) IgUrl {
	return IgBaseURL.append("users/").append(userId).append("/relationship")
}

func UrlMediaSearch() IgUrl {
	return IgBaseURL.append("media/search")
}

func UrlCommentsDelete(mediaId, commentId string) IgUrl {
	return IgBaseURL.append("media/").append(mediaId).append("/comments/").append(commentId)
}

func UrlLocationsMediaRecent(locName string) IgUrl {
	return IgBaseURL.append("locations/").append(locName).append("/media/recent")
}

func UrlLocationsSearch() IgUrl {
	return IgBaseURL.append("locations/search")
}

func UrlGeographiesMediaRecent(geoId string) IgUrl {
	return IgBaseURL.append("geographies/").append(geoId).append("/media/recent")
}
