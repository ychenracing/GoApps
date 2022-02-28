package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	ConcurrentNum = 20                         // 将会开启ConcurrentNum个goroutine来爬取用户相册中所有图片
	SaveFolder    = `E:/Downloads/kongjiewang` // 图片保存的文件夹
)

var startUrl = `http://www.kongjie.com/home.php?mod=space&do=album&view=all&order=hot&page=1`
var headers = map[string][]string{
	"Accept":                    []string{"text/html,application/xhtml+xml,application/xml", "q=0.9,image/webp,*/*;q=0.8"},
	"Accept-Encoding":           []string{"gzip, deflate, sdch"},
	"Accept-Language":           []string{"zh-CN,zh;q=0.8,en;q=0.6,zh-TW;q=0.4"},
	"Accept-Charset":            []string{"utf-8"},
	"Connection":                []string{"keep-alive"},
	"DNT":                       []string{"1"},
	"Host":                      []string{"www.kongjie.com"},
	"Referer":                   []string{"http://www.kongjie.com/home.php?mod=space&do=album&view=all&order=hot&page=1"},
	"Upgrade-Insecure-Requests": []string{"1"},
	"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"},
}

// 用户id和图片id的正则表达式，用于从url中提取用户id和图片id，保存图片时这些id会拼接成图片名
var uidPicIdPattern = regexp.MustCompile(`.*?uid=(\d+).*?picid=(\d+).*?`)

// 图片链接的正则表达式，用于从图片浏览页面的html内容中提取出图片链接，然后保存图片
var imageUrlPattern = regexp.MustCompile(`<div\s+?id="photo_pic"\s+?class="c">(?s:.*?)<a\s+?href=".*?<img\s+?src="(.*?)"\s+?id="pic"`)

// 下一张图片所在的图片浏览页面的链接正则表达式，用于从图片浏览页面提取出下一页链接，翻页爬取
var nextImagePageUrlPattern = regexp.MustCompile(`<div\s+?class="pns\s+?mlnv\s+?vm\s+?mtm\s+?cl">(?s:.*?)<a\s+?href="(.*?)"\s+?class="btn".*?title="下一张"(?s:.*?)<img\s+?src".*?"\s+?alt="下一张"`)

// 用户相册块的正则表达式，用于从相册列表页提取出用户相册块，用户相册块中包含很多个用户的相册链接
var peopleUlPattern = regexp.MustCompile(`<div\s+?class="ptw">(?s:.*?)<ul\s+?class="ml\s+?mlp\s+?cl">(?s:(.*?))</ul>`)

// 用户相册的正则表达式，用于从用户相册块提取出用户相册链接，然后就可以进入相册爬取图片了
var peopleItemPattern = regexp.MustCompile(`<li\s+?class="d">(?s:.*?)<div\s+?class="c">(?s:.*?)<a\s+?href="(.*?)">`)

// 下一个相册列表页链接的正则表达式，用于从相册列表页提取出下一页链接，翻页爬取
var nextAlbumPageUrlPattern = regexp.MustCompile(`<div\s+?class="pgs\s+?cl\s+?mtm">(?s:.*?)</label>(?s:.*?)<a\s+?href="(.*?)"\s+?class="nxt">下一页</a>`)

// redis链接信息
var redisOption = redis.DialPassword("flyvar")                      // redis密码
var redisConn, _ = redis.Dial("tcp", "127.0.0.1:6379", redisOption) // 连接本地redis

// 图片页面通道。每个相册点进去将会进入图片浏览页面，该通道就是为了存放这些图片浏览页面的url，供图片爬取的goroutine使用
var imagePageUrlChan = make(chan string, 200)

var wg sync.WaitGroup

// 串行访问redis，否则goroutine并发访问redis时会报错
var redisLock sync.Mutex

func main() {
	// 创建保存的文件夹
	_, err := os.Open(SaveFolder)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(SaveFolder, 0666)
		}
	}

	// 开启CONCURRENT_NUM个goroutine来爬取用户相册中所有图片的动作
	wg.Add(ConcurrentNum)
	for i := 0; i < ConcurrentNum; i++ {
		go getImagesInAlbum()
	}

	// 开启单个goroutine爬取所有用户的相册链接
	parseAlbumUrl(startUrl)

	// 等待爬取完成
	wg.Wait()
}

// 解析出相册url，然后进入相册爬取图片
func parseAlbumUrl(nextUrl string) {
	for {
		albumHtmlContent := getHtmlFromUrl(nextUrl)

		// FindSubmatch查找正则表达式的匹配和所有的子匹配组，这里是查找当前页每个人的相册链接
		peopleListElement := peopleUlPattern.FindSubmatch(albumHtmlContent)
		if len(peopleListElement) <= 0 {
			// 当前页没有相册
			fmt.Println("no peopleListElement!, url=", nextUrl)
			// 当前页所有用户相册链接解析完毕，翻到下一页
			nextAlbumUrl := nextAlbumPageUrlPattern.FindSubmatch(albumHtmlContent)
			if len(nextAlbumUrl) <= 0 {
				fmt.Println("all albums crawled!")
				break
			}
			nextUrl = string(nextAlbumUrl[1])
			continue
		}

		// 子匹配组是第二个元素。里面包含了很多用户的相册连接
		peopleUlContent := peopleListElement[1]
		peopleItems := peopleItemPattern.FindAllSubmatch(peopleUlContent, -1)
		if len(peopleItems) > 0 {
			for _, peopleItem := range peopleItems {
				if len(peopleItem) <= 0 {
					continue
				}
				// 找到了一个用户的相册链接，放入imagePageUrlChan中等待爬取
				peopleAlbumUrl := strings.ReplaceAll(string(peopleItem[1]), `&amp;`, "&")
				imagePageUrlChan <- peopleAlbumUrl
			}
		}
		// 当前页所有用户相册链接解析完毕，翻到下一页
		nextAlbumUrl := nextAlbumPageUrlPattern.FindSubmatch(albumHtmlContent)
		if len(nextAlbumUrl) <= 0 {
			fmt.Println("all albums crawled!")
			break
		}
		nextUrl = strings.ReplaceAll(string(nextAlbumUrl[1]), `&amp;`, "&")
		fmt.Println(nextUrl)
	}
	close(imagePageUrlChan)
}

// 进入空姐网用户的相册，开始一张一张的保存相册中的图片。
// 解析出uid和picId，用于存储图片的名字
func getImagesInAlbum() {
	for imagePageUrl := range imagePageUrlChan {
		// 从当前图片页面url中获取当前图片所属的用户id和图片id
		uidPicIdMatch := uidPicIdPattern.FindStringSubmatch(imagePageUrl)
		if len(uidPicIdMatch) <= 0 {
			fmt.Println("can not find any uidPicId! imagePageUrl=", imagePageUrl)
			continue
		}
		uid := uidPicIdMatch[1]   // 用户id
		picId := uidPicIdMatch[2] // 图片id

		imagePageHtmlContent := getHtmlFromUrl(imagePageUrl)

		// redis中不存在，说明这张图片没被爬取过
		exists := hexists("kongjie", uid+":"+picId)
		if !exists {
			// 获取图片src，即图片具体链接
			imageSrcList := imageUrlPattern.FindSubmatch(imagePageHtmlContent)
			if len(imageSrcList) > 0 {
				imageSrc := string(imageSrcList[1])
				imageSrc = strings.ReplaceAll(string(imageSrc), `&amp;`, "&")
				saveImage(imageSrc, uid, picId)
				hset("kongjie", uid+":"+picId, "1")
			}
		}

		// 解析下一张图片页面的url，继续爬取
		nextImagePageUrlSubmatch := nextImagePageUrlPattern.FindSubmatch(imagePageHtmlContent)
		if len(nextImagePageUrlSubmatch) <= 0 {
			continue
		}
		nextImagePageUrl := string(nextImagePageUrlSubmatch[1])
		imagePageUrlChan <- nextImagePageUrl
	}
	wg.Done()
}

// 保存图片到全局变量saveFolder文件夹下，图片名字为“uid_picId.ext”。
// 其中，uid是用户id，picId是空姐网图片id，ext是图片的扩展名。
func saveImage(imageUrl string, uid string, picId string) {
	res := getReponseWithGlobalHeaders(imageUrl)
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// 获取图片扩展名
	fileNameExt := path.Ext(imageUrl)
	// 图片保存的全路径
	savePath := path.Join(SaveFolder, uid+"_"+picId+fileNameExt)
	imageWriter, _ := os.OpenFile(savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	length, _ := io.Copy(imageWriter, res.Body)
	fmt.Println(uid + "_" + picId + fileNameExt + " image saved! " + strconv.Itoa(int(length)) + " bytes." + imageUrl)
}

func hexists(key, field string) bool {
	redisLock.Lock()
	defer redisLock.Unlock()
	exists, err := redisConn.Do("HEXISTS", key, field)
	if err != nil {
		fmt.Println("redis hexists error!", err)
	}
	if exists == nil {
		return false
	}
	return exists.(int64) == 1
}

func hset(key, field, value string) bool {
	redisLock.Lock()
	defer redisLock.Unlock()
	ok, err := redisConn.Do("HSET", key, field, value)
	if err != nil {
		fmt.Println("redis hset error!", err)
	}
	if ok == nil {
		return false
	}
	return ok.(int64) == 1
}

func getReponseWithGlobalHeaders(url string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	if headers != nil && len(headers) != 0 {
		for k, v := range headers {
			for _, val := range v {
				req.Header.Add(k, val)
			}
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return res
}

func getHtmlFromUrl(url string) []byte {
	response := getReponseWithGlobalHeaders(url)

	reader := response.Body
	// 返回的内容被压缩成gzip格式了，需要解压一下
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(response.Body)
	}
	// 此时htmlContent还是gbk编码，需要转换成utf8编码
	htmlContent, _ := ioutil.ReadAll(reader)

	oldReader := bufio.NewReader(bytes.NewReader(htmlContent))
	peekBytes, _ := oldReader.Peek(1024)
	e, _, _ := charset.DetermineEncoding(peekBytes, "")
	utf8reader := transform.NewReader(oldReader, e.NewDecoder())
	// 此时htmlContent就已经是utf8编码了
	htmlContent, _ = ioutil.ReadAll(utf8reader)

	if err := response.Body.Close(); err != nil {
		fmt.Println("error happened when closing response body!", err)
	}
	return htmlContent
}
