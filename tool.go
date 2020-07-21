package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Instargam struct {
	Data struct {
		User struct {
			EdgeOwnerToTimelineMedia struct {
				//Count int `json:"count"`
				Edges []struct {
					Node struct {
						/*
							DashInfo struct {
								IsDashEligible    bool   `json:"is_dash_eligible"`
								NumberOfQualities int  `json:"number_of_qualities"`
								VideoDashManifest string `json:"video_dash_manifest"`
							} `json:"dash_info"`

							EdgeMediaToComment struct {
								PageInfo struct {
									EndCursor string `json:"end_cursor"`
								} `json:"page_info"`
							} `json:"edge_media_to_comment"`
						*/
						DisplayURL            string `json:"display_url"`
						EdgeSidecarToChildren struct {
							Edges []struct {
								Node struct {
									/*
										DashInfo struct {
											VideoDashManifest string `json:"video_dash_manifest"`
										} `json:"dash_info"`

									*/
									DisplayURL string `json:"display_url"`
									IsVideo    bool   `json:"is_video"`
									VideoURL   string `json:"video_url"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_sidecar_to_children"`
						IsVideo  bool   `json:"is_video"`
						VideoURL string `json:"video_url"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"end_cursor"`
					HasNextPage bool   `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
}

// 获取网页源代码
func GetHtml(Insurl string) (html string) {
	// 解析代理地址
	proxy, err := url.Parse("http://127.0.0.1:1087") //加载本地代理
	//设置网络传输
	netTransport := &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * time.Duration(5),
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	request, err := http.NewRequest("GET", Insurl, nil)
	if err != nil {
		log.Println(err)
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36") //模拟浏览器User-Agent
	request.Header.Set("Cookie", `ig_did=5448DAB0-E424-48F9-97DF-3A386834D9BE; mid=XsH7PQAEAAG55HTxRZVxBxvZICeY; ds_user_id=25850748707; shbid=2673; shbts=1590030929.7257333; csrftoken=OvLzaajy6C9peOrWQeTfo61GBRY0xLGz; sessionid=25850748707%3ATgB2CXKqw4BFUu%3A23; rur=FTW; urlgen="{\"35.221.222.156\": 15169\054 \"34.92.32.194\": 15169\054 \"34.92.235.238\": 15169}:1jcJdR:nlKH_w6YMze8RyJkjEbQ5Sc9UxA"`)

	res, err := httpClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	//判断是否成功访问，如果成功访问StatusCode应该为200
	if res.StatusCode != http.StatusOK {
		log.Println(err)
		return
	}
	content, _ := ioutil.ReadAll(res.Body)
	return string(content)
}

// 通过主页获取第一个after的值
func GetAfterByHtml(homepage string) string {
	regex := `end_cursor":"(.*?)"},"edges`
	rp := regexp.MustCompile(regex)
	after := rp.FindStringSubmatch(homepage)
	return after[1]
}

// 通过主页获得博主的账号ID
func GetIdByHtml(homepage string) string {
	regex := `{"id":"(.*?)","username`
	rp := regexp.MustCompile(regex)
	id := rp.FindStringSubmatch(homepage)
	return id[1]
}

//获取博主的账号名
func GetUserName(homepage string) string {
	l1 := len("https://www.instagram.com/")
	l2 := len(homepage)
	return homepage[l1:l2-1] + ".txt"
}

// 拼接查询地址
func SetQueryUrl(query_hash, id, first, after string) string {
	url := fmt.Sprintf("https://www.instagram.com/graphql/query/?query_hash=%s&variables={\"id\":\"%s\",\"first\":%s,\"after\":\"%s\"}", query_hash, id, first, after)
	return url
}

//json转Struct
func Json2Struct(strjson string) Instargam {
	var ins Instargam
	json.Unmarshal([]byte(strjson), &ins)
	return ins
}

//通过json得到After
func GetAfter(ins Instargam) string {
	return ins.Data.User.EdgeOwnerToTimelineMedia.PageInfo.EndCursor
}

// 判断网站是否加载到底
func IsEnd(ins Instargam) bool {
	return ins.Data.User.EdgeOwnerToTimelineMedia.PageInfo.HasNextPage
}

// 通过json得到图片和视频下载地址
func GetDownloadUrl(savefile string, ins Instargam) {
	for _, v := range ins.Data.User.EdgeOwnerToTimelineMedia.Edges {
		var content string
		if v.Node.IsVideo != true {
			fmt.Println(v.Node.DisplayURL)
			content = v.Node.DisplayURL + "\n"
		} else {
			fmt.Println(v.Node.VideoURL)
			content += v.Node.VideoURL + "\n"
		}
		for _, v1 := range v.Node.EdgeSidecarToChildren.Edges {
			if v1.Node.IsVideo != true {
				fmt.Println(v1.Node.DisplayURL)
				content = v1.Node.DisplayURL + "\n"
			} else {
				fmt.Println(v1.Node.VideoURL)
				content += v1.Node.VideoURL + "\n"
			}
		}
		WirteText(savefile, content)
	}
}

// 写入txt文件
func WirteText(savefile string, txt string) {
	f, err := os.OpenFile(savefile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("os Create error: ", err)
		return
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	bw.WriteString(txt)
	bw.Flush()
}

func main() {
	var homepage string
	fmt.Scanln(&homepage)
	query_hash := "44efc15d3c13342d02df0b5a9fa3d33f"
	first := 12
	//通过html获取第一条json地址
	html := GetHtml(homepage)
	after := GetAfterByHtml(html)
	id := GetIdByHtml(html)
	first_query_url := SetQueryUrl(query_hash, id, strconv.Itoa(first), after) //  第一条json地址

	//通过json获取内容
	jsonconent := GetHtml(first_query_url)
	for {
		ins := Json2Struct(jsonconent)
		next_after := GetAfter(ins) //通过json获得after的值
		next_query_url := SetQueryUrl(query_hash, id, strconv.Itoa(first), next_after)
		GetDownloadUrl(GetUserName(homepage), ins)
		jsonconent = GetHtml(next_query_url)

		if !IsEnd(ins) { //如果页面加载到底，结束循环
			break
		}
	}

}