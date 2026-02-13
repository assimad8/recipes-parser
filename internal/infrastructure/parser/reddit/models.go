package reddit

type Listing struct {
	Data ListingData `json:"data"`
}

type ListingData struct {
	Children []Post `json:"children"`
}

type Post struct {
	Kind string   `json:"kind"`
	Data PostData `json:"data"`
}

type PostData struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	CreatedUTC    float64 `json:"created_utc"`
	URL           string  `json:"url"`
	URLOverride   string  `json:"url_overridden_by_dest"`
	Permalink     string  `json:"permalink"`
	LinkFlairText string  `json:"link_flair_text"`
	Thumbnail     string  `json:"thumbnail"`
}