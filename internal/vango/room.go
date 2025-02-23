package vango

import (
	"net/http"
	"strconv"
	"time"
)

type Room struct {
	Hash string
}

func NewRoom(hash string) Room {
	return Room{
		Hash: hash,
	}
}

func (room *Room) Join(client *Client) (statusCode int) {
	url := "https://draw.kuku.lu/index.php?action=addRoomJson&_=" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	request, _ := http.NewRequest("GET", url, nil)
	response, _ := client.HttpClient.Do(request)
	client.Status = "Joined"
	return response.StatusCode
}
