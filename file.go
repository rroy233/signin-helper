package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	Logger "github.com/rroy233/logger"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var cosClient *cos.Client

type FileOptions struct {
	AllowContentType []string `json:"allow_content_type"`
	MaxSize          int      `json:"max_size"`
	Rename           bool     `json:"rename"`
}

var fileExt = map[string]string{
	"image/png":                    ".png",
	"image/jpeg":                   ".jpg",
	"application/zip":              ".zip",
	"application/x-zip-compressed": ".zip",
	"application/x-rar-compressed": ".rar",
	"application/msword":           ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel": ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.ms-powerpoint":                                             ".ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
	"application/pdf": ".pdf",
}

func cosClientUpdate() {
	u, _ := url.Parse(config.Cos.BucketUrl)
	su, _ := url.Parse("https://cos.COS_REGION.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	// 1.永久密钥
	cosClient = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.Cos.SecretID,
			SecretKey: config.Cos.SecretKey,
		},
	})
}

func cosFileExist(objKey string) bool {
	resp, err := cosClient.Object.Head(ctx, objKey, nil)
	if err != nil {
		Logger.Error.Println("[COS]获取文件元数据失败:", err)
		return false
	}
	contentLength := resp.Header.Get("Content-Length")

	if contentLength == "0" {
		Logger.Debug.Println("[COS]文件不存在" + objKey)
		return false
	}
	return true
}

func cosFileDel(objKey string) error {
	_, err := cosClient.Object.Delete(context.Background(), objKey)
	return err
}

func cosUpload(localAddr string, fileName string) (ObjectKey string, err error) {
	_, err = os.Open(localAddr)
	if err != nil {
		Logger.Info.Println("[COS]尝试上传本地文件时，打开本地文件失败:", err)
		return "", err
	}

	upRes, _, err := cosClient.Object.Upload(context.Background(), config.Cos.BasePath+"/upload/"+fileName, localAddr, nil)
	if err != nil || upRes == nil {
		Logger.Error.Println("[COS]尝试上传本地文件时失败:", err)
		return "", err
	}
	Logger.Info.Printf("[COS]上传文件成功:%s->%s\n", localAddr, upRes.Key)
	return upRes.Key, err
}

func cosDownload(objKey string, dest string) error {
	opt := &cos.MultiDownloadOptions{
		ThreadPoolSize: 5,
	}
	_, err := cosClient.Object.Download(
		context.Background(), objKey, dest, opt,
	)
	return err
}

func cosGetUrl(objKey string, expTime time.Duration, imageCompress bool) (imgUrl string, err error) {
	var tmp *url.URL
	tmp, err = cosClient.Object.GetPresignedURL(ctx, "GET", objKey, config.Cos.SecretID, config.Cos.SecretKey, expTime, nil)
	if err != nil {
		Logger.Error.Println("[COS]获取签名url失败:", err.Error())
		return
	}

	fileUrl := tmp.String()
	if imageCompress == true {
		fileUrl += "&imageMogr2/format/jpg/interlace/0/quality/36"
	}
	fileToken, err := Cipher.Encrypt([]byte(fileUrl))
	if err != nil {
		Logger.Error.Println("[COS]生成代理url失败:", err.Error())
		return
	}
	imgUrl = config.General.BaseUrl + "/file/" + fileToken + "." + MD5(fileToken+config.General.AESKey)
	return
}

// Compress 压缩文件
//files 文件数组，可以是不同dir下的文件或者文件夹
//dest 压缩文件存放地址
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compress(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// FsExists 判断文件/文件夹是否存在
func FsExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fileHandler(c *gin.Context) {
	p := c.Param("data")

	//校验
	data := strings.Split(p, ".")
	if len(data) != 2 {
		returnErrorText(c, 403, "参数无效")
		return
	}
	if data[1] != MD5(data[0]+config.General.AESKey) {
		returnErrorText(c, 403, "链接无效")
		return
	}

	dc, err := Cipher.Decrypt(data[0])
	if err != nil {
		returnErrorText(c, 403, "链接无效")
		return
	}

	fileNotFoundImage, _ := ioutil.ReadFile("./static/image/image_fileNotFound.jpg")

	/*协议
	redis_tempImage/*  => 从redis提取
	local_file#*       => 本地文件
	其他                => 使用http.get获取
	*/
	if strings.Contains(dc, "redis_tempImage/") == true { //从redis提取
		tmpImageToken := strings.Split(dc, "/")[1]
		rdbData, err := rdb.Get(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", tmpImageToken)).Result()
		if err != nil {
			c.Data(404, "image/jpeg", fileNotFoundImage)
			Logger.Info.Printf("[文件代理] 从redis提取 - 读取rdb失败：%s", err.Error())
			return
		}
		fileData, err := base64.StdEncoding.DecodeString(rdbData)
		if err != nil {
			c.Data(404, "image/jpeg", fileNotFoundImage)
			Logger.Info.Printf("[文件代理] 从redis提取 - base64解码失败：%s", err.Error())
			return
		}
		c.Data(200, "image/jpeg", fileData)
	} else if strings.Contains(dc, "local_file#") == true {
		fileContentType := strings.Split(dc, "#")[1]
		LocalFile := strings.Split(dc, "#")[2]
		fileData, err := ioutil.ReadFile(LocalFile)
		if err != nil {
			c.Data(404, "image/jpeg", fileNotFoundImage)
			Logger.Info.Printf("[文件代理] 从本地读取 - readFile失败：%s", err.Error())
			return
		}
		c.Data(200, fileContentType, fileData)
	} else { //使用http.get获取
		resp, err := http.Get(dc)
		if err != nil {
			c.Status(resp.StatusCode)
			Logger.Info.Printf("[文件代理]请求错误：code[%d]，header[%v]", resp.StatusCode, resp.Header)
			return
		}
		defer resp.Body.Close()

		//连接过期
		if resp.StatusCode != 200 {
			c.Data(resp.StatusCode, "image/jpeg", fileNotFoundImage)
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)

		//附件
		if strings.Contains(resp.Header.Get("content-type"), "image") != true {
			urlData, err := url.Parse(dc)
			if err != nil {
				returnErrorText(c, 500, "文件解析失败")
				Logger.Error.Println("[文件代理]文件解析失败:", err)
				return
			}
			c.Header("content-type", resp.Header.Get("content-type"))
			//获取文件名
			pathData := strings.Split(urlData.Path, "/")
			fileName := "file"
			if len(pathData) != 0 {
				fileName = pathData[len(pathData)-1]
			}
			c.Header("Content-Disposition", "attachment;filename="+fileName)
		}

		c.Data(200, resp.Header.Get("content-type"), body)
	}
	return
}
