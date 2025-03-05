package vango

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn        *websocket.Conn
	HttpClient  http.Client
	Id          string
	KeepRoutine bool
	Status      string
	Ucode       string
}

type OpenRoomResponse struct {
	Result string
	Hash   string
}

func NewClient(useragent string) Client {
	jar, _ := cookiejar.New(nil)

	client := http.Client{Jar: jar}
	request, _ := http.NewRequest("GET", "https://draw.kuku.lu/", nil)
	request.Header.Add("User-Agent", useragent)
	client.Do(request)

	cookies := jar.Cookies(request.URL)

	const letters = "0123456789abcdef"
	brand := make([]byte, 5)
	rand.Read(brand)
	var id string
	for _, v := range brand {
		id += string(letters[int(v)%len(letters)])
	}

	return Client{
		HttpClient: client,
		Id:         id,
		Status:     "Ready",
		Ucode:      cookies[0].Value,
	}
}

func (client *Client) OpenRoom() (hash string) {
	url := fmt.Sprintf("https://draw.kuku.lu/index.php?action=addRoomJson&_=%s", strconv.FormatInt(time.Now().UnixMilli(), 10))
	request, _ := http.NewRequest("GET", url, nil)
	response, _ := client.HttpClient.Do(request)

	var _response OpenRoomResponse
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &_response)
	client.Status = "Joined"
	return _response.Hash
}

func (client *Client) JoinRoom(hash string, name string, hostaddr string) (statusCode int) {
	url := fmt.Sprintf("https://draw.kuku.lu/pchat.php?action=enterRoomJson&hash=%s&joinname=%s&entermode=&_=%s", hash, name, strconv.FormatInt(time.Now().UnixMilli(), 10))
	request, _ := http.NewRequest("GET", url, nil)
	response, _ := client.HttpClient.Do(request)

	conn, _, _ := websocket.DefaultDialer.Dial(hostaddr, nil)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("@{\"type\":\"join2\",\"uucode\":\"%s\",\"useragent\":\"Chrome\",\"hash\":\"%s\"}", client.Ucode, hash)))
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("@{\"type\":\"config_user\",\"id\":\"%s\",\"color\":\"#000000\",\"mouse_mode\":-1,\"wis\":\"\"}", client.Id)))

	client.StartRoutine()

	client.Status = "Joined"
	client.Conn = conn
	return response.StatusCode
}

func (client *Client) Move(x int, y int) {
	client.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("0/%s/%d/%d////////0/", client.Id, x, y)))
}

func (client *Client) Draw(x int, y int, color string) {
	client.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("0/%s/%d/%d/%d/%d/%s/2.0/100/1/0/", client.Id, x, y, x, y, color)))
}

func (client *Client) Erase(x int, y int) {

}

func (client *Client) StartRoutine() {
	client.KeepRoutine = true
	for {
		if !client.KeepRoutine {
			break
		}

		_, message, _ := client.Conn.ReadMessage()
		if message != nil {
			fmt.Println(string(message))
		}
	}
}

func (client *Client) StopRoutine() {
	client.KeepRoutine = false
}
