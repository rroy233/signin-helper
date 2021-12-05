package main

import (
	"archive/zip"
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
	"os"
	"signin/Logger"
	"time"
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
	"application/x-rar-compressed": ".rar",
	"application/msword":           ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel": ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.ms-powerpoint":                                             ".ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
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

func cosGetUrl(objKey string, expTime time.Duration) (imgUrl string, err error) {
	var tmp *url.URL
	tmp, err = cosClient.Object.GetPresignedURL(ctx, "GET", objKey, config.Cos.SecretID, config.Cos.SecretKey, expTime, nil)
	if err != nil {
		Logger.Error.Println("[COS]获取签名url失败:", err.Error())
		return
	}
	imgUrl = tmp.String()
	Logger.Info.Println("[COS]生成签名URL成功:", tmp.String())
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
