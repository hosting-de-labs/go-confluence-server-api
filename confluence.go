package confluence

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL   *url.URL
	UserAgent string

	Username string
	Password string

	HttpClient *http.Client
}

type ProfilePicture struct {
	Path      string `json:"path,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

type By struct {
	Type           string          `json:"type,omitempty"`
	Username       string          `json:"username,omitempty"`
	UserKey        string          `json:"userKey,omitempty"`
	ProfilePicture *ProfilePicture `json:"profilePicture,omitempty"`
	DisplayName    string          `json:"displayName,omitempty"`
	Links          *Links          `json:"_links,omitempty"`
	Expendable     *Expendable     `json:"_expandable,omitempty"`
}

type History struct {
	Latest    bool `json:"latest,omitempty"`
	CreatedBy *By  `json:"createdBy,omitempty"`
}

type Expendable struct {
	Ancestors       string `json:"ancestors,omitempty"`
	Children        string `json:"children,omitempty"`
	Container       string `json:"container,omitempty"`
	Content         string `json:"content,omitempty"`
	History         string `json:"history,omitempty"`
	Metadata        string `json:"metadata,omitempty"`
	Icon            string `json:"icon,omitempty"`
	Space           string `json:"space,omitempty"`
	Description     string `json:"description,omitempty"`
	Homepage        string `json:"homepage,omitempty"`
	Version         string `json:"version,omitempty"`
	LastUpdated     string `json:"lastUpdated,omitempty"`
	PreviousVersion string `json:"previousVersion,omitempty"`
	Contributors    string `json:"contributors,omitempty"`
	NextVersion     string `json:"nextVersion,omitempty"`
}

type Links struct {
	Base       string `json:"base,omitempty"`
	Collection string `json:"collection,omitempty"`
	Self       string `json:"self,omitempty"`
	Tinyui     string `json:"tinyui,omitempty"`
	Webui      string `json:"webui,omitempty"`
}

type Storage struct {
	Representation string      `json:"representation,omitempty"`
	Value          string      `json:"value,omitempty"`
	Expendable     *Expendable `json:"_expandable,omitempty"`
}

type BodyPart struct {
	Expendable *Expendable `json:"_expandable,omitempty"`
}

type Body struct {
	Editor         *BodyPart `json:"editor,omitempty"`
	Representation string    `json:"representation,omitempty"`
	Value          string    `json:"value,omitempty"`
	Storage        *Storage  `json:"storage,omitempty"`
}

type Space struct {
	ID         int         `json:"id,omitempty"`
	Key        string      `json:"key,omitempty"`
	Name       string      `json:"name,omitempty"`
	Type       string      `json:"type,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expendable *Expendable `json:"_expandable,omitempty"`
}

type Version struct {
	By         *By         `json:"by,omitempty"`
	When       string      `json:"when,omitempty"`
	Message    string      `json:"message,omitempty"`
	Number     int         `json:"number,omitempty"`
	MinorEdit  bool        `json:"minorEdit,omitempty"`
	Hidden     bool        `json:"hidden,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expendable *Expendable `json:"_expandable,omitempty"`
}

type Ancestor struct {
	ID string `json:"id,omitempty"`
}

type Page struct {
	Expendable *Expendable `json:"_expandable,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	ID         string      `json:"id,omitempty"`
	Title      string      `json:"title,omitempty"`
	Type       string      `json:"type,omitempty"`
	Body       *Body       `json:"body,omitempty"`
	Status     string      `json:"status,omitempty"`
	Space      *Space      `json:"space,omitempty"`
	Version    *Version    `json:"version,omitempty"`
	Ancestors  []*Ancestor `json:"ancestors,omitempty"`
}

/*
Get
curl -u admin:admin http://localhost:8080/confluence/rest/api/content/64819161?expand=body.storage |
python -mjson.tool

Create
curl -u admin:admin -X POST -H 'Content-Type: application/json' -d '{"type":"page","title":"new page",
"ancestors":[{"id":456}], "space":{"key":"TST"},"body":{"storage":{"value":
"<p>This is a new page</p>","representation":"storage"}}}'
http://localhost:8080/confluence/rest/api/content/ | python -mjson.tool

Update
curl -u admin:admin -X PUT -H 'Content-Type: application/json' -d '{"id":"3604482","type":"page",
"title":"new page","space":{"key":"TST"},"body":{"storage":{"value":
"<p>This is the updated text for the new page</p>","representation":"storage"}},
"version":{"number":2}}' http://localhost:8080/confluence/rest/api/content/3604482 | python -mjson.tool
*/

func (c *Client) CreatePage(space string, parentId string, title string, body string) (Page, error) {
	var ancestor Ancestor
	ancestor.ID = parentId

	var page Page
	page.Ancestors = append(page.Ancestors, &ancestor)
	page.Title = title
	page.Type = "page"

	var spaceObject Space
	spaceObject.Key = space
	page.Space = &spaceObject

	var storageObject Storage
	storageObject.Value = body
	storageObject.Representation = "storage"

	var bodyObject Body
	bodyObject.Storage = &storageObject
	page.Body = &bodyObject

	var versionObject Version
	versionObject.Number = 1
	page.Version = &versionObject

	req, err := c.newRequest("POST", "/content/", page)
	if err != nil {
		return page, err
	}
	_, err = c.do(req, &page)

	return page, err
}

func (c *Client) UpdatePage(space string, id string, title string, body string, version int) (Page, error) {
	var page Page
	page.ID = id
	page.Title = title
	page.Type = "page"

	var spaceObject Space
	spaceObject.Key = space
	page.Space = &spaceObject

	var storageObject Storage
	storageObject.Value = body
	storageObject.Representation = "storage"

	var bodyObject Body
	bodyObject.Storage = &storageObject
	page.Body = &bodyObject

	var versionObject Version
	versionObject.Number = version
	page.Version = &versionObject

	req, err := c.newRequest("PUT", "/content/"+id, page)
	if err != nil {
		return page, err
	}
	_, err = c.do(req, &page)

	return page, err
}

func (c *Client) GetPage(id string) (Page, error) {
	var page Page
	req, err := c.newRequest("GET", "/content/"+id+"?expand=body.storage,version", nil)
	if err != nil {
		return page, err
	}
	_, err = c.do(req, &page)
	return page, err
}

func NewAPI(location string, username string, password string) (*Client, error) {
	if len(location) == 0 || len(username) == 0 || len(password) == 0 {
		return nil, errors.New("url, username or password empty")
	}

	u, err := url.ParseRequestURI(location)

	if err != nil {
		return nil, err
	}

	c := new(Client)
	httpClient := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}
	c.HttpClient = &httpClient
	c.BaseURL = u
	c.Password = password
	c.Username = username
	c.UserAgent = "confluence server API Client"

	return c, nil
}

func (c *Client) basicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.BaseURL.String()+path, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	req.Header.Add("Authorization", "Basic "+c.basicAuth(c.Username, c.Password))
	resp, err := c.HttpClient.Do(req)

	if err != nil {
		return nil, err
	}

	// TODO status codes abfangen

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)

	return resp, err
}
