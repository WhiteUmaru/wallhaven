package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/reujab/wallpaper"
)

const SEARCH_LIST = "https://wallhaven.cc/api/v1/search?apikey=4APAKnqFKolSF2wCnMcQqId4Mx5CXEOQ"

var (
	categories   = "110"
	purity       = "110"
	sorting      = "random"
	order        = "desc"
	topRange     = "1M"
	pager        = 1
	__dirPath, _ = os.Getwd()
	client       = http.Client{}
	wg           sync.WaitGroup
)

const MAX_POOL_SIZE = 4

func main() {
	fmt.Println("欢迎使用壁纸引擎")
	fmt.Println("请输入模式：1.默认模式 2.自定义模式 3. 批量下载模式")
	model := "1"
	fmt.Scanln(&model)
	if model == "2" {
		fmt.Println("开始自定义模式")
		setConfig()
	} else if model == "3" {
		setConfig()
		setDownloadType()
		fmt.Println("开始下载图片列表")
		pool := tunny.NewFunc(MAX_POOL_SIZE, func(obj interface{}) interface{} {
			switch photo := obj.(type) {
			case PhotoInfo:
				fmt.Println("开始下载照片" + photo.Id)
				file := downloadImage(photo.Path)
				fmt.Println("下载成功！图片保存路径为：" + file)
				time.Sleep(time.Second)
			}
			return nil
		})
		defer pool.Close()

		startDownloadImage(pool)
		wg.Wait()
		return
	}

	file := getRandomImage()
	wg.Wait()
	wallpaper.SetFromFile(file)
	fmt.Println("再见~")
}

func setConfig() {
	fmt.Println("请设置风格：100/101/111* /etc (general/anime/people)")
	var mCategories string
	fmt.Scanln(&mCategories)
	if mCategories != "" {
		categories = mCategories
	}
	fmt.Println("请设置图片级别：100/101/111* /etc (sfw/sketchy/nsfw)")
	fmt.Scanln(&purity)
}

func setDownloadType() {
	sorting = "toplist"
	order = "desc"
	fmt.Println("请输入要下载的页数：default：1")
	fmt.Scanln(&pager)
}

func startDownloadImage(pool *tunny.Pool) {
	if pager <= 0 {
		fmt.Println("列表获取完成")
		return
	}
	data := requestImageList().Data
	pager--
	wg.Add(len(data))
	fmt.Println("当前获取到的图片数量" + strconv.Itoa(len(data)))
	for _, datum := range data {
		go pool.Process(datum)
	}
	startDownloadImage(pool)
}

func getRandomImage() string {
	wg.Add(1)
	fmt.Println("开始请求照片列表")
	data := requestImageList().Data
	if len(data) == 0 {
		return getRandomImage()
	}
	path := data[0].Path
	if path == "" {
		return getRandomImage()
	}
	fmt.Println("开始下载照片" + data[0].Id)
	file := downloadImage(path)
	fmt.Println("下载成功！图片保存路径为：" + file)
	return file
}

func downloadImage(url string) string {
	defer wg.Done()
	resp, err := client.Get(url + "?apikey=4APAKnqFKolSF2wCnMcQqId4Mx5CXEOQ")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	strs := strings.Split(url, "/")

	filePath := __dirPath + "/image/" + strs[len(strs)-1]
	os.MkdirAll(__dirPath+"/image/", 0777)
	// 如果文件已存在 则返回
	if Exists(filePath) {
		return filePath
	}
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	write.Write(body)
	write.Flush()
	if FileSize(filePath) < 1024*10 {
		fmt.Println("下载失败:" + strs[len(strs)-1])
		os.Remove(filePath)
		return ""
	}
	return filePath
}

func requestImageList() SearchResponse {
	resp, err := client.Get(SEARCH_LIST + "&categories=" + categories + "&purity=" + purity + "&sorting=" + sorting + "&page=" + strconv.Itoa(pager) + "&order=" + order + "&topRange=" + topRange)
	result := SearchResponse{}
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	return result
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err) || false
	}
	return true
}
func FileSize(path string) int64 {
	info, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return 0
	}
	return info.Size()
}

type PhotoInfo struct {
	Path string `json:"path"`
	Id   string `json:"id"`
}
type SearchResponse struct {
	Data []PhotoInfo `json:"data"`
}
